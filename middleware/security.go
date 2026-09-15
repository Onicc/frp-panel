package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Security-Policy", "default-src 'self'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; connect-src 'self' ws: wss:")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2<<20)
		}
		c.Next()
	}
}

type rateVisitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func LoginRateLimit() gin.HandlerFunc {
	var mu sync.Mutex
	visitors := make(map[string]*rateVisitor)
	return func(c *gin.Context) {
		host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			host = c.Request.RemoteAddr
		}
		now := time.Now()
		mu.Lock()
		for key, visitor := range visitors {
			if now.Sub(visitor.lastSeen) > 15*time.Minute {
				delete(visitors, key)
			}
		}
		visitor := visitors[host]
		if visitor == nil {
			visitor = &rateVisitor{limiter: rate.NewLimiter(rate.Every(12*time.Second), 5)}
			visitors[host] = visitor
		}
		visitor.lastSeen = now
		allowed := visitor.limiter.Allow()
		mu.Unlock()
		if !allowed {
			c.Header("Retry-After", "12")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": http.StatusTooManyRequests, "msg": "too many login attempts"})
			return
		}
		c.Next()
	}
}

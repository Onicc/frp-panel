package v2

import (
	"net/http"
	"time"

	protocol "github.com/Onicc/frp-panel/internal/protocol/v2"
	"github.com/Onicc/frp-panel/middleware"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/gin-gonic/gin"
)

type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

func Configure(router *gin.RouterGroup, appInstance app.Application) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC(), "protocol": protocol.Version})
	})
	router.POST("/agent/enroll", middleware.LoginRateLimit(), redeemEnrollment(appInstance))
	protected := router.Group("", middleware.JWTAuth(appInstance), middleware.AuthCtx(appInstance), middleware.RBAC(appInstance))
	protected.GET("/capabilities", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"controller": gin.H{"os": "linux", "docker": true, "embeddedFrps": true},
			"agent":      protocol.CurrentCapabilities(),
		})
	})
	protected.POST("/enrollments", createEnrollment(appInstance))
}

func AbortProblem(c *gin.Context, status int, title, detail string) {
	c.Header("Content-Type", "application/problem+json")
	c.AbortWithStatusJSON(status, Problem{
		Type:  "https://github.com/Onicc/frp-panel/blob/main/docs/api-problems.md",
		Title: title, Status: status, Detail: detail, Instance: c.Request.URL.Path,
	})
}

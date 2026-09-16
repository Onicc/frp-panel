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
	router.GET("/bootstrap-status", bootstrapStatus(appInstance))
	router.POST("/agent/enroll", middleware.LoginRateLimit(), redeemEnrollment(appInstance))
	router.POST("/server/enroll", middleware.LoginRateLimit(), redeemServerEnrollment(appInstance))
	protected := router.Group("", middleware.JWTAuth(appInstance), middleware.AuthCtx(appInstance), middleware.RBAC(appInstance))
	protected.GET("/capabilities", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"master": gin.H{"os": "linux", "docker": true, "embeddedFrps": false, "standaloneFrps": true},
			"agent":  protocol.CurrentCapabilities(),
		})
	})
	protected.GET("/account", getAccount(appInstance))
	protected.POST("/account/password", changePassword(appInstance))
	protected.POST("/enrollments", createEnrollment(appInstance))
	protected.POST("/server-enrollments", createServerEnrollment(appInstance))
	protected.GET("/tunnels", listTunnels(appInstance))
	protected.POST("/tunnels", createTunnel(appInstance))
	protected.DELETE("/tunnels", deleteTunnel(appInstance))
}

func AbortProblem(c *gin.Context, status int, title, detail string) {
	c.Header("Content-Type", "application/problem+json")
	c.AbortWithStatusJSON(status, Problem{
		Type:  "https://github.com/Onicc/frp-panel/blob/main/docs/api-problems.md",
		Title: title, Status: status, Detail: detail, Instance: c.Request.URL.Path,
	})
}

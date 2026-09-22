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
	Type         string `json:"type"`
	Title        string `json:"title"`
	Status       int    `json:"status"`
	Detail       string `json:"detail,omitempty"`
	Instance     string `json:"instance,omitempty"`
	Dependencies any    `json:"dependencies,omitempty"`
}

func Configure(router *gin.RouterGroup, appInstance app.Application) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().UTC(), "protocol": protocol.Version})
	})
	router.GET("/bootstrap-status", bootstrapStatus(appInstance))
	router.POST("/agent/enroll", middleware.LoginRateLimit(), redeemEnrollment(appInstance))
	router.POST("/agent/location", reportClientLocation(appInstance))
	router.POST("/server/enroll", middleware.LoginRateLimit(), redeemServerEnrollment(appInstance))
	router.POST("/auth/login", middleware.LoginRateLimit(), login(appInstance))
	router.POST("/auth/register", middleware.LoginRateLimit(), register(appInstance))
	protected := router.Group("", middleware.JWTAuth(appInstance), middleware.AuthCtx(appInstance), middleware.RBAC(appInstance))
	protected.GET("/capabilities", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"master": gin.H{"os": "linux", "docker": true, "embeddedFrps": false, "standaloneFrps": true},
			"agent":  protocol.CurrentCapabilities(),
		})
	})
	protected.GET("/account", getAccount(appInstance))
	protected.GET("/account/session", getSessionAccount(appInstance))
	protected.POST("/auth/logout", logout(appInstance))
	protected.POST("/account/password", changePassword(appInstance))
	protected.GET("/overview", overview(appInstance))
	protected.GET("/updates/release", releaseStatus(appInstance))
	protected.POST("/updates/master", startMasterUpdate(appInstance))
	protected.GET("/updates/operations/:id", updateOperation(appInstance))
	protected.GET("/topology", topology(appInstance))
	protected.GET("/clients", listClients(appInstance))
	protected.POST("/clients", createClient(appInstance))
	protected.GET("/clients/:id", getClient(appInstance))
	protected.PATCH("/clients/:id", patchClient(appInstance))
	protected.DELETE("/clients/:id", deleteClient(appInstance))
	protected.POST("/clients/:id/enrollment", rotateClientEnrollment(appInstance))
	protected.GET("/servers", listServers(appInstance))
	protected.POST("/servers", createServer(appInstance))
	protected.GET("/servers/:id", getServer(appInstance))
	protected.PATCH("/servers/:id", patchServer(appInstance))
	protected.DELETE("/servers/:id", deleteServer(appInstance))
	protected.POST("/servers/:id/enrollment", rotateServerEnrollment(appInstance))
	protected.PATCH("/servers/:id/update-policy", serverUpdatePolicy(appInstance))
	protected.POST("/servers/:id/updates", startServerUpdateHandler(appInstance))
	protected.GET("/tunnels", listTunnelsResource(appInstance))
	protected.POST("/tunnels", createTunnelResource(appInstance))
	protected.GET("/tunnels/:id", getTunnel(appInstance))
	protected.PATCH("/tunnels/:id", patchTunnel(appInstance))
	protected.DELETE("/tunnels/:id", deleteTunnelResource(appInstance))
	// Legacy enrollment endpoints remain available for existing automation.
	protected.POST("/enrollments", createEnrollment(appInstance))
	protected.POST("/server-enrollments", createServerEnrollment(appInstance))
	protected.DELETE("/tunnels", deleteTunnel(appInstance))
}

func AbortProblem(c *gin.Context, status int, title, detail string) {
	c.Header("Content-Type", "application/problem+json")
	c.AbortWithStatusJSON(status, Problem{
		Type:  "https://github.com/Onicc/frp-panel/blob/main/docs/api-problems.md",
		Title: title, Status: status, Detail: detail, Instance: c.Request.URL.Path,
	})
}

func AbortProblemWithDependencies(c *gin.Context, status int, title, detail string, dependencies any) {
	c.Header("Content-Type", "application/problem+json")
	c.AbortWithStatusJSON(status, Problem{Type: "https://github.com/Onicc/frp-panel/blob/main/docs/api-problems.md", Title: title, Status: status, Detail: detail, Instance: c.Request.URL.Path, Dependencies: dependencies})
}

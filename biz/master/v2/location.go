package v2

import (
	"net/http"
	"strings"
	"time"

	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils"
	"github.com/gin-gonic/gin"
)

// reportClientLocation uses the Agent's long-lived credential, independently
// of the browser session. The connection peer IP remains stored separately.
func reportClientLocation(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := strings.TrimSpace(c.GetHeader("X-FRP-Panel-Client-ID"))
		secret := c.GetHeader("X-FRP-Panel-Client-Secret")
		if clientID == "" || secret == "" {
			AbortProblem(c, http.StatusUnauthorized, "Unauthorized", "Agent credentials are required")
			return
		}
		var item models.Client
		db := appInstance.GetDBManager().GetDefaultDB()
		if db.Where("client_id = ? AND (origin_client_id IS NULL OR origin_client_id = ?)", clientID, "").First(&item).Error != nil || !utils.CheckCredential(secret, item.ConnectSecret) {
			AbortProblem(c, http.StatusUnauthorized, "Unauthorized", "invalid Agent credential")
			return
		}
		var body struct {
			IP string `json:"ip"`
		}
		if c.ShouldBindJSON(&body) != nil || publicLocationIP(body.IP) == "" {
			AbortProblem(c, http.StatusBadRequest, "Invalid IP", "a public IPv4 or IPv6 address is required")
			return
		}
		now := time.Now().UTC()
		if err := db.Model(&models.Client{}).Where("client_id = ?", clientID).Updates(map[string]any{"reported_ip": publicLocationIP(body.IP), "reported_at": now}).Error; err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Location report failed", "could not save Client location")
			return
		}
		c.Status(http.StatusNoContent)
	}
}

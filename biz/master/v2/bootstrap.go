package v2

import (
	"net/http"

	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/dao"
	"github.com/gin-gonic/gin"
)

type bootstrapStatusResponse struct {
	RegistrationEnabled bool `json:"registrationEnabled"`
	OwnerExists         bool `json:"ownerExists"`
	CanCreateOwner      bool `json:"canCreateOwner"`
}

// bootstrapStatus exposes only the state needed to render the first-run form.
// It intentionally returns no user records or tenant information.
func bootstrapStatus(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		count, err := dao.NewQuery(app.NewContext(c, appInstance)).AdminCountUsers()
		if err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Bootstrap status unavailable", "could not inspect owner setup state")
			return
		}
		enabled := appInstance.GetConfig().App.EnableRegister
		c.JSON(http.StatusOK, bootstrapStatusResponse{
			RegistrationEnabled: enabled,
			OwnerExists:         count > 0,
			CanCreateOwner:      enabled && count == 0,
		})
	}
}

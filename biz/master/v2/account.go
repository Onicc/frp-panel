package v2

import (
	"net/http"

	"github.com/Onicc/frp-panel/common"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/dao"
	"github.com/Onicc/frp-panel/utils"
	"github.com/gin-gonic/gin"
)

type accountResponse struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required"`
}

func getAccount(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInfo := common.GetUserInfo(c)
		if userInfo == nil || !userInfo.Valid() {
			AbortProblem(c, http.StatusUnauthorized, "Unauthorized", "a valid user session is required")
			return
		}
		user, err := dao.NewQuery(app.NewContext(c, appInstance)).GetUserByUserID(userInfo.GetUserID())
		if err != nil {
			AbortProblem(c, http.StatusNotFound, "Account not found", "the signed-in account no longer exists")
			return
		}
		c.JSON(http.StatusOK, accountResponse{Username: user.UserName, Email: user.Email, Role: user.Role})
	}
}

func changePassword(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request changePasswordRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			AbortProblem(c, http.StatusBadRequest, "Invalid password request", "currentPassword and newPassword are required")
			return
		}
		if len(request.NewPassword) < 12 {
			AbortProblem(c, http.StatusBadRequest, "Password is too short", "the new password must contain at least 12 characters")
			return
		}
		if request.CurrentPassword == request.NewPassword {
			AbortProblem(c, http.StatusBadRequest, "Password is unchanged", "choose a password different from the current password")
			return
		}
		userInfo := common.GetUserInfo(c)
		if userInfo == nil || !userInfo.Valid() {
			AbortProblem(c, http.StatusUnauthorized, "Unauthorized", "a valid user session is required")
			return
		}
		ctx := app.NewContext(c, appInstance)
		user, err := dao.NewQuery(ctx).GetUserByUserID(userInfo.GetUserID())
		if err != nil || !utils.CheckPasswordHash(request.CurrentPassword, user.Password) {
			AbortProblem(c, http.StatusUnauthorized, "Current password is incorrect", "enter the current account password and try again")
			return
		}
		hash, err := utils.HashPassword(request.NewPassword)
		if err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Password change failed", "the new password could not be secured")
			return
		}
		user.Password = hash
		if err := dao.NewMutation(ctx).UpdateUser(userInfo, user); err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Password change failed", "the account could not be updated")
			return
		}
		cfg := appInstance.GetConfig()
		c.SetCookie(cfg.App.CookieName, "", -1, cfg.App.CookiePath, cfg.App.CookieDomain, cfg.App.CookieSecure, cfg.App.CookieHTTPOnly)
		c.JSON(http.StatusOK, gin.H{"reauthenticate": true})
	}
}

package v2

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/Onicc/frp-panel/common"
	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/dao"
	"github.com/Onicc/frp-panel/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type authRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type safeAccount struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func accountFromUser(user models.UserInfo) safeAccount {
	return safeAccount{Username: user.GetUserName(), Email: user.GetEmail(), Role: user.GetRole()}
}

func login(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request authRequest
		if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Username) == "" || request.Password == "" {
			AbortProblem(c, http.StatusBadRequest, "Invalid sign-in", "username and password are required")
			return
		}
		ctx := app.NewContext(c, appInstance)
		ok, user, err := dao.NewQuery(ctx).CheckUserPassword(strings.TrimSpace(request.Username), request.Password)
		if err != nil || !ok || user == nil || !user.Valid() {
			AbortProblem(c, http.StatusUnauthorized, "Sign-in failed", "the username or password is incorrect")
			return
		}
		token, err := conf.GetJWTWithPayload(appInstance.GetConfig(), user.GetUserID(), map[string]interface{}{})
		if err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Sign-in failed", "could not create a secure session")
			return
		}
		middlewarePushToken(c, appInstance, token)
		c.JSON(http.StatusOK, gin.H{"user": accountFromUser(user)})
	}
}

func register(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request authRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			AbortProblem(c, http.StatusBadRequest, "Invalid registration", "username, email, and password are required")
			return
		}
		request.Username = strings.TrimSpace(request.Username)
		request.Email = strings.TrimSpace(request.Email)
		parsedEmail, emailErr := mail.ParseAddress(request.Email)
		if !utils.IsClientIDPermited(request.Username) || len(request.Password) < 12 || emailErr != nil || parsedEmail.Address != request.Email {
			AbortProblem(c, http.StatusBadRequest, "Invalid registration", "use a URL-safe username, a valid email, and a password of at least 12 characters")
			return
		}
		ctx := app.NewContext(c, appInstance)
		count, err := dao.NewQuery(ctx).AdminCountUsers()
		if err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Registration unavailable", "could not inspect account setup")
			return
		}
		if !appInstance.GetConfig().App.EnableRegister || count > 0 {
			AbortProblem(c, http.StatusConflict, "Registration unavailable", "owner registration has already been completed")
			return
		}
		if err := dao.NewQuery(ctx).CheckUserNameAndEmail(request.Username, request.Email); err == nil {
			AbortProblem(c, http.StatusConflict, "Account already exists", "choose a different username or email")
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			AbortProblem(c, http.StatusInternalServerError, "Registration unavailable", "could not validate account uniqueness")
			return
		}
		hash, err := utils.HashPassword(request.Password)
		if err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Registration failed", "could not secure the password")
			return
		}
		user := &models.UserEntity{UserName: request.Username, Email: request.Email, Password: hash, Status: models.STATUS_NORMAL, Role: defs.UserRole_Owner, Token: uuid.NewString()}
		if err := dao.NewMutation(ctx).CreateUser(user); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") {
				AbortProblem(c, http.StatusConflict, "Account already exists", "choose a different username or email")
				return
			}
			AbortProblem(c, http.StatusInternalServerError, "Registration failed", "could not create the owner account")
			return
		}
		token, err := conf.GetJWTWithPayload(appInstance.GetConfig(), user.UserID, map[string]interface{}{})
		if err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Registration failed", "could not create a secure session")
			return
		}
		middlewarePushToken(c, appInstance, token)
		c.JSON(http.StatusCreated, gin.H{"user": accountFromUser(user)})
	}
}

func logout(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := appInstance.GetConfig()
		c.SetSameSite(http.SameSiteStrictMode)
		c.SetCookie(cfg.App.CookieName, "", -1, cfg.App.CookiePath, cfg.App.CookieDomain, cfg.App.CookieSecure, cfg.App.CookieHTTPOnly)
		c.Status(http.StatusNoContent)
	}
}

// Kept as a tiny wrapper to make auth handlers easy to test without importing
// the legacy auth package.
func middlewarePushToken(c *gin.Context, appInstance app.Application, token string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(appInstance.GetConfig().App.CookieName, token, appInstance.GetConfig().App.CookieAge, appInstance.GetConfig().App.CookiePath, appInstance.GetConfig().App.CookieDomain, appInstance.GetConfig().App.CookieSecure, appInstance.GetConfig().App.CookieHTTPOnly)
}

func getSessionAccount(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := common.GetUserInfo(c)
		if user == nil || !user.Valid() {
			AbortProblem(c, http.StatusUnauthorized, "Unauthorized", "a valid user session is required")
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": accountFromUser(user)})
	}
}

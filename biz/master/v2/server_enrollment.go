package v2

import (
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Onicc/frp-panel/common"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type createServerEnrollmentRequest struct {
	ServerID string `json:"serverId" binding:"required"`
	ServerIP string `json:"serverIp" binding:"required"`
	BindPort int    `json:"bindPort"`
}

type createServerEnrollmentResponse struct {
	ServerID  string    `json:"serverId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type redeemServerEnrollmentResponse struct {
	ServerID string `json:"serverId"`
	Secret   string `json:"secret"`
}

func createServerEnrollment(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request createServerEnrollmentRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			AbortProblem(c, http.StatusBadRequest, "Invalid request", "serverId and serverIp are required")
			return
		}
		request.ServerID = strings.TrimSpace(request.ServerID)
		request.ServerIP = strings.TrimSpace(request.ServerIP)
		if !utils.IsClientIDPermited(request.ServerID) {
			AbortProblem(c, http.StatusBadRequest, "Invalid server ID", "use letters, numbers, underscores, or hyphens")
			return
		}
		if !validServerHost(request.ServerIP) {
			AbortProblem(c, http.StatusBadRequest, "Invalid server address", "use a public IP address or DNS hostname without a scheme or port")
			return
		}
		if request.BindPort == 0 {
			request.BindPort = 7000
		}
		if request.BindPort < 1024 || request.BindPort > 65535 {
			AbortProblem(c, http.StatusBadRequest, "Invalid bind port", "bindPort must be between 1024 and 65535")
			return
		}

		userInfo := common.GetUserInfo(c)
		if userInfo == nil || !userInfo.Valid() {
			AbortProblem(c, http.StatusUnauthorized, "Unauthorized", "a valid user session is required")
			return
		}
		token, err := randomToken()
		if err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Enrollment failed", "could not generate a secure token")
			return
		}
		globalID := app.GlobalClientID(userInfo.GetUserName(), "s", request.ServerID)
		expiresAt := time.Now().UTC().Add(enrollmentLifetime)
		secret := utils.DeriveCredential(appInstance.GetConfig().App.GlobalSecret, "frps-server", token)
		server := &models.ServerEntity{
			ServerID: globalID, TenantID: userInfo.GetTenantID(), UserID: userInfo.GetUserID(),
			ServerIP: request.ServerIP, ConnectSecret: utils.HashCredential(secret),
		}
		if err := server.SetConfigContent(utils.NewBaseFRPServerUserAuthConfig(request.BindPort, nil)); err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Enrollment failed", "could not create the initial FRPS configuration")
			return
		}

		db := appInstance.GetDBManager().GetDefaultDB()
		err = db.Transaction(func(tx *gorm.DB) error {
			var previous models.ServerEnrollment
			previousError := tx.Where("server_id = ?", globalID).First(&previous).Error
			if previousError == nil {
				if previous.UsedAt != nil {
					return gorm.ErrDuplicatedKey
				}
				if err := tx.Delete(&previous).Error; err != nil {
					return err
				}
				if err := tx.Unscoped().Where("server_id = ? AND user_id = ? AND tenant_id = ?", globalID, userInfo.GetUserID(), userInfo.GetTenantID()).Delete(&models.Server{}).Error; err != nil {
					return err
				}
			} else if !errors.Is(previousError, gorm.ErrRecordNotFound) {
				return previousError
			}
			if err := tx.Create(&models.Server{ServerEntity: server}).Error; err != nil {
				return err
			}
			return tx.Create(&models.ServerEnrollment{
				ID: uuid.NewString(), TokenHash: utils.HashCredential(token), ServerID: globalID,
				UserID: userInfo.GetUserID(), TenantID: userInfo.GetTenantID(), ExpiresAt: expiresAt,
			}).Error
		})
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "unique") {
				AbortProblem(c, http.StatusConflict, "Server already exists", "choose a different server ID")
				return
			}
			AbortProblem(c, http.StatusInternalServerError, "Enrollment failed", "the enrollment could not be stored")
			return
		}
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusCreated, createServerEnrollmentResponse{ServerID: globalID, Token: token, ExpiresAt: expiresAt})
	}
}

func redeemServerEnrollment(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request redeemEnrollmentRequest
		if err := c.ShouldBindJSON(&request); err != nil || len(request.Token) < 32 || len(request.Token) > 256 {
			AbortProblem(c, http.StatusBadRequest, "Invalid request", "a valid enrollment token is required")
			return
		}

		var response redeemServerEnrollmentResponse
		now := time.Now().UTC()
		db := appInstance.GetDBManager().GetDefaultDB()
		err := db.Transaction(func(tx *gorm.DB) error {
			var enrollment models.ServerEnrollment
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("token_hash = ? AND used_at IS NULL", utils.HashCredential(request.Token)).
				First(&enrollment).Error; err != nil {
				return err
			}
			if !now.Before(enrollment.ExpiresAt) {
				return errEnrollmentExpired
			}
			var server models.Server
			if err := tx.Where("server_id = ?", enrollment.ServerID).First(&server).Error; err != nil {
				return err
			}
			secret := utils.DeriveCredential(appInstance.GetConfig().App.GlobalSecret, "frps-server", request.Token)
			if !utils.CheckCredential(secret, server.ConnectSecret) {
				return errEnrollmentInvalid
			}
			result := tx.Model(&models.ServerEnrollment{}).
				Where("id = ? AND used_at IS NULL", enrollment.ID).Update("used_at", now)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return errEnrollmentInvalid
			}
			response = redeemServerEnrollmentResponse{ServerID: enrollment.ServerID, Secret: secret}
			return nil
		})
		if err != nil {
			AbortProblem(c, http.StatusUnauthorized, "Enrollment rejected", "the token is invalid, expired, or already used")
			return
		}
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusOK, response)
	}
}

func validServerHost(value string) bool {
	if value == "" || len(value) > 253 || strings.ContainsAny(value, "/\\?#@ \t\r\n") {
		return false
	}
	if net.ParseIP(value) != nil {
		return true
	}
	if strings.Contains(value, ":") {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, character := range label {
			if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') &&
				(character < '0' || character > '9') && character != '-' {
				return false
			}
		}
	}
	return true
}

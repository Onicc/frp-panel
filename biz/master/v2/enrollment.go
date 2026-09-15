package v2

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
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

const enrollmentLifetime = 10 * time.Minute

type createEnrollmentRequest struct {
	NodeID string `json:"nodeId" binding:"required"`
}

type createEnrollmentResponse struct {
	NodeID    string    `json:"nodeId"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type redeemEnrollmentRequest struct {
	Token string `json:"token" binding:"required"`
}

type redeemEnrollmentResponse struct {
	NodeID string `json:"nodeId"`
	Secret string `json:"secret"`
}

func createEnrollment(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request createEnrollmentRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			AbortProblem(c, http.StatusBadRequest, "Invalid request", "nodeId is required")
			return
		}
		request.NodeID = strings.TrimSpace(request.NodeID)
		if !utils.IsClientIDPermited(request.NodeID) {
			AbortProblem(c, http.StatusBadRequest, "Invalid node ID", "use letters, numbers, underscores, or hyphens")
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
		globalID := app.GlobalClientID(userInfo.GetUserName(), "c", request.NodeID)
		expiresAt := time.Now().UTC().Add(enrollmentLifetime)
		secret := utils.DeriveCredential(appInstance.GetConfig().App.GlobalSecret, "agent-node", token)
		db := appInstance.GetDBManager().GetDefaultDB()
		err = db.Transaction(func(tx *gorm.DB) error {
			client := &models.Client{ClientEntity: &models.ClientEntity{
				ClientID: globalID, TenantID: userInfo.GetTenantID(), UserID: userInfo.GetUserID(),
				ConnectSecret: utils.HashCredential(secret), IsShadow: true,
			}}
			if err := tx.Create(client).Error; err != nil {
				return err
			}
			return tx.Create(&models.AgentEnrollment{
				ID: uuid.NewString(), TokenHash: utils.HashCredential(token), ClientID: globalID,
				UserID: userInfo.GetUserID(), TenantID: userInfo.GetTenantID(), ExpiresAt: expiresAt,
			}).Error
		})
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "unique") {
				AbortProblem(c, http.StatusConflict, "Node already exists", "choose a different node ID")
				return
			}
			AbortProblem(c, http.StatusInternalServerError, "Enrollment failed", "the enrollment could not be stored")
			return
		}
		c.Header("Cache-Control", "no-store")
		c.JSON(http.StatusCreated, createEnrollmentResponse{NodeID: globalID, Token: token, ExpiresAt: expiresAt})
	}
}

func redeemEnrollment(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request redeemEnrollmentRequest
		if err := c.ShouldBindJSON(&request); err != nil || len(request.Token) < 32 || len(request.Token) > 256 {
			AbortProblem(c, http.StatusBadRequest, "Invalid request", "a valid enrollment token is required")
			return
		}

		var response redeemEnrollmentResponse
		now := time.Now().UTC()
		db := appInstance.GetDBManager().GetDefaultDB()
		err := db.Transaction(func(tx *gorm.DB) error {
			var enrollment models.AgentEnrollment
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("token_hash = ? AND used_at IS NULL", utils.HashCredential(request.Token)).
				First(&enrollment).Error; err != nil {
				return err
			}
			if !now.Before(enrollment.ExpiresAt) {
				return errEnrollmentExpired
			}
			var client models.Client
			if err := tx.Where("client_id = ?", enrollment.ClientID).First(&client).Error; err != nil {
				return err
			}
			secret := utils.DeriveCredential(appInstance.GetConfig().App.GlobalSecret, "agent-node", request.Token)
			if !utils.CheckCredential(secret, client.ConnectSecret) {
				return errEnrollmentInvalid
			}
			result := tx.Model(&models.AgentEnrollment{}).
				Where("id = ? AND used_at IS NULL", enrollment.ID).Update("used_at", now)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return errEnrollmentInvalid
			}
			response = redeemEnrollmentResponse{NodeID: enrollment.ClientID, Secret: secret}
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

var (
	errEnrollmentExpired = errors.New("enrollment expired")
	errEnrollmentInvalid = errors.New("enrollment invalid")
)

func randomToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

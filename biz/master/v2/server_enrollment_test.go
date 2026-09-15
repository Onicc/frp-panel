package v2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestServerEnrollmentIsOneUseAndStoresOnlyHashes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	application := app.NewApp()
	cfg := conf.DefaultConfig()
	cfg.App.GlobalSecret = strings.Repeat("a", 32)
	application.SetConfig(cfg)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	dbManager := models.NewDBManager(defs.DBTypeSQLite3)
	dbManager.SetDB(defs.DBTypeSQLite3, defs.DBRoleDefault, db)
	dbManager.Init()
	application.SetDBManager(dbManager)

	user := &models.UserEntity{UserID: 7, TenantID: 3, UserName: "owner", Status: models.STATUS_NORMAL}
	createRouter := gin.New()
	createRouter.Use(func(c *gin.Context) { c.Set(defs.UserInfoKey, user); c.Next() })
	createRouter.POST("/api/v2/server-enrollments", createServerEnrollment(application))

	request := httptest.NewRequest(http.MethodPost, "/api/v2/server-enrollments", bytes.NewBufferString(`{"serverId":"edge","serverIp":"edge.example.test","bindPort":7100}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	createRouter.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var enrollment createServerEnrollmentResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &enrollment); err != nil {
		t.Fatal(err)
	}
	if enrollment.ServerID != "owner.s.edge" || len(enrollment.Token) < 32 {
		t.Fatalf("unexpected enrollment: %#v", enrollment)
	}

	var stored models.Server
	if err := db.Where("server_id = ?", enrollment.ServerID).First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	secret := utils.DeriveCredential(cfg.App.GlobalSecret, "frps-server", enrollment.Token)
	if stored.ConnectSecret == secret || !utils.CheckCredential(secret, stored.ConnectSecret) {
		t.Fatal("server credential was not stored as a one-way hash")
	}
	serverConfig, err := stored.GetConfigContent()
	if err != nil || serverConfig.BindPort != 7100 {
		t.Fatalf("unexpected server config: %#v, %v", serverConfig, err)
	}
	oldToken := enrollment.Token
	replacementRecorder := httptest.NewRecorder()
	replacementRequest := httptest.NewRequest(http.MethodPost, "/api/v2/server-enrollments", bytes.NewBufferString(`{"serverId":"edge","serverIp":"edge.example.test","bindPort":7100}`))
	replacementRequest.Header.Set("Content-Type", "application/json")
	createRouter.ServeHTTP(replacementRecorder, replacementRequest)
	if replacementRecorder.Code != http.StatusCreated {
		t.Fatalf("replace unused enrollment status = %d, body = %s", replacementRecorder.Code, replacementRecorder.Body.String())
	}
	if err := json.Unmarshal(replacementRecorder.Body.Bytes(), &enrollment); err != nil {
		t.Fatal(err)
	}
	if enrollment.Token == oldToken {
		t.Fatal("replacement enrollment reused the old token")
	}

	redeemRouter := gin.New()
	redeemRouter.POST("/api/v2/server/enroll", redeemServerEnrollment(application))
	oldRedeemBody, _ := json.Marshal(map[string]string{"token": oldToken})
	oldRecorder := httptest.NewRecorder()
	redeemRouter.ServeHTTP(oldRecorder, httptest.NewRequest(http.MethodPost, "/api/v2/server/enroll", bytes.NewReader(oldRedeemBody)))
	if oldRecorder.Code != http.StatusUnauthorized {
		t.Fatalf("replaced token status = %d, body = %s", oldRecorder.Code, oldRecorder.Body.String())
	}
	redeemBody, _ := json.Marshal(map[string]string{"token": enrollment.Token})
	for attempt, expectedStatus := range []int{http.StatusOK, http.StatusUnauthorized} {
		redeemRecorder := httptest.NewRecorder()
		redeemRouter.ServeHTTP(redeemRecorder, httptest.NewRequest(http.MethodPost, "/api/v2/server/enroll", bytes.NewReader(redeemBody)))
		if redeemRecorder.Code != expectedStatus {
			t.Fatalf("redeem attempt %d status = %d, body = %s", attempt+1, redeemRecorder.Code, redeemRecorder.Body.String())
		}
	}
}

func TestValidServerHost(t *testing.T) {
	for _, value := range []string{"edge.example.com", "203.0.113.10", "2001:db8::1"} {
		if !validServerHost(value) {
			t.Fatalf("valid host rejected: %q", value)
		}
	}
	for _, value := range []string{"", "https://edge.example.com", "edge.example.com:7000", "edge.example.com/path", "edge example.com"} {
		if validServerHost(value) {
			t.Fatalf("invalid host accepted: %q", value)
		}
	}
}

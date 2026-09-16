package v2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
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

func TestAccountPasswordChangeRequiresCurrentPasswordAndEndsSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	application := app.NewApp()
	cfg := conf.DefaultConfig()
	cfg.App.CookieName = "frp_panel_session"
	application.SetConfig(cfg)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "account.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	dbManager := models.NewDBManager(defs.DBTypeSQLite3)
	dbManager.SetDB(defs.DBTypeSQLite3, defs.DBRoleDefault, db)
	dbManager.Init()
	application.SetDBManager(dbManager)
	hash, err := utils.HashPassword("current-password")
	if err != nil {
		t.Fatal(err)
	}
	user := &models.UserEntity{
		UserID: 9, TenantID: 5, UserName: "owner", Email: "owner@example.test",
		Password: hash, Role: defs.UserRole_Owner, Status: models.STATUS_NORMAL,
	}
	if err := db.Create(&models.User{UserEntity: user}).Error; err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(defs.UserInfoKey, user); c.Next() })
	router.GET("/api/v2/account", getAccount(application))
	router.POST("/api/v2/account/password", changePassword(application))

	profile := httptest.NewRecorder()
	router.ServeHTTP(profile, httptest.NewRequest(http.MethodGet, "/api/v2/account", nil))
	if profile.Code != http.StatusOK || !strings.Contains(profile.Body.String(), `"username":"owner"`) || strings.Contains(profile.Body.String(), "Password") {
		t.Fatalf("account response = %d %s", profile.Code, profile.Body.String())
	}

	change := func(current, next string) *httptest.ResponseRecorder {
		t.Helper()
		body, _ := json.Marshal(changePasswordRequest{CurrentPassword: current, NewPassword: next})
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v2/account/password", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, request)
		return recorder
	}
	if recorder := change("wrong-password", "a-new-password-123"); recorder.Code != http.StatusUnauthorized {
		t.Fatalf("wrong current password status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if recorder := change("current-password", "too-short"); recorder.Code != http.StatusBadRequest {
		t.Fatalf("short new password status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	recorder := change("current-password", "a-new-password-123")
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Header().Get("Set-Cookie"), "frp_panel_session=") || !strings.Contains(recorder.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Fatalf("successful password change = %d headers=%v body=%s", recorder.Code, recorder.Header(), recorder.Body.String())
	}
	var stored models.User
	if err := db.Where("user_id = ?", user.UserID).First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if !utils.CheckPasswordHash("a-new-password-123", stored.Password) || utils.CheckPasswordHash("current-password", stored.Password) {
		t.Fatal("stored password was not replaced securely")
	}
}

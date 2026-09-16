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

func TestAgentEnrollmentCanReplaceOnlyAnUnusedToken(t *testing.T) {
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
	createRouter.POST("/api/v2/enrollments", createEnrollment(application))
	create := func() (int, createEnrollmentResponse) {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v2/enrollments", bytes.NewBufferString(`{"nodeId":"mac"}`))
		request.Header.Set("Content-Type", "application/json")
		createRouter.ServeHTTP(recorder, request)
		var response createEnrollmentResponse
		_ = json.Unmarshal(recorder.Body.Bytes(), &response)
		return recorder.Code, response
	}

	status, first := create()
	if status != http.StatusCreated || first.NodeID != "owner.c.mac" || len(first.Token) < 32 {
		t.Fatalf("first enrollment status = %d, response = %#v", status, first)
	}
	status, replacement := create()
	if status != http.StatusCreated || replacement.Token == first.Token {
		t.Fatalf("replacement enrollment status = %d, response = %#v", status, replacement)
	}

	redeemRouter := gin.New()
	redeemRouter.POST("/api/v2/agent/enroll", redeemEnrollment(application))
	redeem := func(token string) int {
		body, _ := json.Marshal(map[string]string{"token": token})
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v2/agent/enroll", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		redeemRouter.ServeHTTP(recorder, request)
		return recorder.Code
	}
	if status := redeem(first.Token); status != http.StatusUnauthorized {
		t.Fatalf("replaced token status = %d", status)
	}
	if status := redeem(replacement.Token); status != http.StatusOK {
		t.Fatalf("replacement token status = %d", status)
	}
	if status, _ := create(); status != http.StatusConflict {
		t.Fatalf("redeemed node recreation status = %d", status)
	}

	var client models.Client
	if err := db.Where("client_id = ?", replacement.NodeID).First(&client).Error; err != nil {
		t.Fatal(err)
	}
	secret := utils.DeriveCredential(cfg.App.GlobalSecret, "agent-node", replacement.Token)
	if client.ConnectSecret == secret || !utils.CheckCredential(secret, client.ConnectSecret) {
		t.Fatal("node credential was not stored as a one-way hash")
	}
}

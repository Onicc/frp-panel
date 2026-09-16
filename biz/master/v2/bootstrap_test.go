package v2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestBootstrapStatusFollowsRegistrationAndOwnerState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	application := app.NewApp()
	cfg := conf.DefaultConfig()
	cfg.App.EnableRegister = true
	application.SetConfig(cfg)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	dbManager := models.NewDBManager(defs.DBTypeSQLite3)
	dbManager.SetDB(defs.DBTypeSQLite3, defs.DBRoleDefault, db)
	dbManager.Init()
	application.SetDBManager(dbManager)

	router := gin.New()
	router.GET("/api/v2/bootstrap-status", bootstrapStatus(application))
	read := func() bootstrapStatusResponse {
		t.Helper()
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v2/bootstrap-status", nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
		}
		var response bootstrapStatusResponse
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		return response
	}

	before := read()
	if !before.RegistrationEnabled || before.OwnerExists || !before.CanCreateOwner {
		t.Fatalf("unexpected initial bootstrap state: %#v", before)
	}
	if err := db.Create(&models.User{UserEntity: &models.UserEntity{
		UserName: "owner", Email: "owner@example.test", Password: "hash", Role: defs.UserRole_Owner,
	}}).Error; err != nil {
		t.Fatal(err)
	}
	after := read()
	if !after.RegistrationEnabled || !after.OwnerExists || after.CanCreateOwner {
		t.Fatalf("unexpected completed bootstrap state: %#v", after)
	}
}

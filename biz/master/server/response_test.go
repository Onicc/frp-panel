package server

import (
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestServerInventoryDoesNotExposeRuntimeConfiguration(t *testing.T) {
	application := app.NewApp()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "servers.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	dbManager := models.NewDBManager(defs.DBTypeSQLite3)
	dbManager.SetDB(defs.DBTypeSQLite3, defs.DBRoleDefault, db)
	dbManager.Init()
	application.SetDBManager(dbManager)
	user := &models.UserEntity{UserID: 3, TenantID: 2, UserName: "owner", Status: models.STATUS_NORMAL}
	server := &models.Server{ServerEntity: &models.ServerEntity{
		ServerID: "owner.s.edge", UserID: user.UserID, TenantID: user.TenantID,
		ServerIP: "edge.example.test", ConfigContent: []byte(`{"auth":{"token":"must-not-leak"}}`),
	}}
	if err := db.Create(server).Error; err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	ginContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginContext.Set(defs.UserInfoKey, user)
	ctx := app.NewContext(ginContext, application)

	listed, err := ListServersHandler(ctx, &pb.ListServersRequest{Page: serverInt32Pointer(1), PageSize: serverInt32Pointer(20)})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.GetServers()) != 1 || listed.GetServers()[0].Config != nil {
		t.Fatalf("Server inventory exposed runtime config: %#v", listed.GetServers())
	}

	got, err := GetServerHandler(ctx, &pb.GetServerRequest{ServerId: serverStringPointer(server.ServerID)})
	if err != nil {
		t.Fatal(err)
	}
	if got.GetServer().Config != nil {
		t.Fatalf("Server detail exposed runtime config: %#v", got.GetServer())
	}
}

func serverInt32Pointer(value int32) *int32    { return &value }
func serverStringPointer(value string) *string { return &value }

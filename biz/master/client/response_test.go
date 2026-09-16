package client

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

func TestClientInventoryDoesNotExposeRuntimeConfiguration(t *testing.T) {
	application := app.NewApp()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "clients.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	dbManager := models.NewDBManager(defs.DBTypeSQLite3)
	dbManager.SetDB(defs.DBTypeSQLite3, defs.DBRoleDefault, db)
	dbManager.Init()
	application.SetDBManager(dbManager)
	user := &models.UserEntity{UserID: 3, TenantID: 2, UserName: "owner", Status: models.STATUS_NORMAL}
	client := &models.Client{ClientEntity: &models.ClientEntity{
		ClientID: "owner.c.mac", UserID: user.UserID, TenantID: user.TenantID,
		IsShadow: true, ConfigContent: []byte(`{"metadatas":{"frp-panel-token":"must-not-leak"}}`),
	}}
	if err := db.Create(client).Error; err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	ginContext, _ := gin.CreateTestContext(httptest.NewRecorder())
	ginContext.Set(defs.UserInfoKey, user)
	ctx := app.NewContext(ginContext, application)

	listed, err := ListClientsHandler(ctx, &pb.ListClientsRequest{Page: int32Pointer(1), PageSize: int32Pointer(20)})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed.GetClients()) != 1 || listed.GetClients()[0].Config != nil {
		t.Fatalf("Client inventory exposed runtime config: %#v", listed.GetClients())
	}

	got, err := GetClientHandler(ctx, &pb.GetClientRequest{ClientId: stringPointer(client.ClientID)})
	if err != nil {
		t.Fatal(err)
	}
	if got.GetClient().Config != nil {
		t.Fatalf("Client detail exposed runtime config: %#v", got.GetClient())
	}
}

func int32Pointer(value int32) *int32    { return &value }
func stringPointer(value string) *string { return &value }

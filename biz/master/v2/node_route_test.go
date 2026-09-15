package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	masterclient "github.com/Onicc/frp-panel/biz/master/client"
	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/rpc"
	"github.com/Onicc/frp-panel/utils"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestNodeRouteCreatesDiscoverableManagedFRPC(t *testing.T) {
	gin.SetMode(gin.TestMode)
	application := app.NewApp()
	cfg := conf.DefaultConfig()
	cfg.App.GlobalSecret = strings.Repeat("b", 32)
	application.SetConfig(cfg)
	application.SetClientsManager(rpc.NewClientsManager())
	application.SetClientRecvMap(&sync.Map{})
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	dbManager := models.NewDBManager(defs.DBTypeSQLite3)
	dbManager.SetDB(defs.DBTypeSQLite3, defs.DBRoleDefault, db)
	dbManager.Init()
	application.SetDBManager(dbManager)

	user := &models.UserEntity{UserID: 8, TenantID: 4, UserName: "owner", Token: "frps-user-token", Status: models.STATUS_NORMAL}
	server := &models.Server{ServerEntity: &models.ServerEntity{
		ServerID: "owner.s.edge", UserID: user.UserID, TenantID: user.TenantID,
		ServerIP: "edge.example.test", ConnectSecret: utils.HashCredential("server-secret"),
	}}
	if err := server.SetConfigContent(utils.NewBaseFRPServerUserAuthConfig(7200, nil)); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(server).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Client{ClientEntity: &models.ClientEntity{
		ClientID: "owner.c.node", UserID: user.UserID, TenantID: user.TenantID,
		ConnectSecret: utils.HashCredential("client-secret"), IsShadow: true,
	}}).Error; err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(defs.UserInfoKey, user); c.Next() })
	router.POST("/api/v2/node-routes", createNodeRoute(application))
	router.GET("/api/v2/node-routes", listNodeRoutes(application))
	router.DELETE("/api/v2/node-routes", deleteNodeRoute(application))
	body, _ := json.Marshal(createNodeRouteRequest{NodeID: "owner.c.node", ServerID: "owner.s.edge"})

	createRecorder := httptest.NewRecorder()
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v2/node-routes", bytes.NewReader(body))
	createRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(createRecorder, createRequest)
	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("create route status = %d, body = %s", createRecorder.Code, createRecorder.Body.String())
	}

	pulled, err := masterclient.RPCPullConfig(app.NewContext(context.Background(), application), &pb.PullClientConfigReq{
		Base: &pb.ClientBase{ClientId: "owner.c.node", ClientSecret: "client-secret"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(pulled.GetClient().GetClientIds()) != 1 {
		t.Fatalf("managed node did not discover its FRPC route: %#v", pulled.GetClient().GetClientIds())
	}
	var child models.Client
	if err := db.Where("client_id = ?", pulled.GetClient().GetClientIds()[0]).First(&child).Error; err != nil {
		t.Fatal(err)
	}
	clientConfig, err := child.GetConfigContent()
	if err != nil || child.ServerID != "owner.s.edge" || clientConfig.ServerAddr != "edge.example.test" || clientConfig.ServerPort != 7200 {
		t.Fatalf("unexpected managed FRPC route: child=%#v config=%#v err=%v", child.ClientEntity, clientConfig, err)
	}

	listRecorder := httptest.NewRecorder()
	router.ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/api/v2/node-routes", nil))
	if listRecorder.Code != http.StatusOK || !strings.Contains(listRecorder.Body.String(), "owner.s.edge") {
		t.Fatalf("list routes status = %d, body = %s", listRecorder.Code, listRecorder.Body.String())
	}

	deleteRecorder := httptest.NewRecorder()
	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v2/node-routes", bytes.NewReader(body))
	deleteRequest.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(deleteRecorder, deleteRequest)
	if deleteRecorder.Code != http.StatusNoContent {
		t.Fatalf("delete route status = %d, body = %s", deleteRecorder.Code, deleteRecorder.Body.String())
	}
	var count int64
	if err := db.Model(&models.Client{}).Where("origin_client_id = ?", "owner.c.node").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("route child remains after delete: count=%d err=%v", count, err)
	}
}

package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sort"
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

func TestTunnelsAutomaticallyManageClientServerConnections(t *testing.T) {
	gin.SetMode(gin.TestMode)
	application := app.NewApp()
	cfg := conf.DefaultConfig()
	cfg.App.GlobalSecret = strings.Repeat("b", 32)
	application.SetConfig(cfg)
	application.SetClientsManager(rpc.NewClientsManager())
	application.SetClientRecvMap(&sync.Map{})
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tunnels.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	dbManager := models.NewDBManager(defs.DBTypeSQLite3)
	dbManager.SetDB(defs.DBTypeSQLite3, defs.DBRoleDefault, db)
	dbManager.Init()
	application.SetDBManager(dbManager)

	user := &models.UserEntity{UserID: 8, TenantID: 4, UserName: "owner", Token: "frps-user-token", Status: models.STATUS_NORMAL}
	for _, definition := range []struct {
		id   string
		host string
		port int
	}{{"owner.s.edge-a", "edge-a.example.test", 7200}, {"owner.s.edge-b", "edge-b.example.test", 7300}} {
		server := &models.Server{ServerEntity: &models.ServerEntity{
			ServerID: definition.id, UserID: user.UserID, TenantID: user.TenantID,
			ServerIP: definition.host, ConnectSecret: utils.HashCredential("server-secret"),
		}}
		if err := server.SetConfigContent(utils.NewBaseFRPServerUserAuthConfig(definition.port, nil)); err != nil {
			t.Fatal(err)
		}
		if err := db.Create(server).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Create(&models.Client{ClientEntity: &models.ClientEntity{
		ClientID: "owner.c.mac", UserID: user.UserID, TenantID: user.TenantID,
		ConnectSecret: utils.HashCredential("client-secret"), IsShadow: true,
	}}).Error; err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.Use(func(c *gin.Context) { c.Set(defs.UserInfoKey, user); c.Next() })
	router.POST("/api/v2/tunnels", createTunnel(application))
	router.GET("/api/v2/tunnels", listTunnels(application))
	router.DELETE("/api/v2/tunnels", deleteTunnel(application))

	create := func(name, serverID string, remotePort int) tunnelResponse {
		t.Helper()
		body, _ := json.Marshal(createTunnelRequest{
			Name: name, ClientID: "owner.c.mac", ServerID: serverID, Type: "tcp",
			LocalHost: "127.0.0.1", LocalPort: 22, RemotePort: remotePort,
		})
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v2/tunnels", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusCreated {
			t.Fatalf("create tunnel %s status = %d, body = %s", name, recorder.Code, recorder.Body.String())
		}
		var response tunnelResponse
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.ClientID != "owner.c.mac" || strings.Contains(recorder.Body.String(), "@") {
			t.Fatalf("tunnel exposed an internal connection: %s", recorder.Body.String())
		}
		return response
	}
	remove := func(name, serverID string) {
		t.Helper()
		body, _ := json.Marshal(deleteTunnelRequest{Name: name, ClientID: "owner.c.mac", ServerID: serverID})
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodDelete, "/api/v2/tunnels", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("delete tunnel %s status = %d, body = %s", name, recorder.Code, recorder.Body.String())
		}
	}
	childCount := func(serverID string) int64 {
		t.Helper()
		var count int64
		if err := db.Model(&models.Client{}).Where("origin_client_id = ? AND server_id = ?", "owner.c.mac", serverID).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		return count
	}

	create("ssh", "owner.s.edge-a", 6022)
	create("web", "owner.s.edge-a", 8080)
	create("dns", "owner.s.edge-b", 8053)
	if got := childCount("owner.s.edge-a"); got != 1 {
		t.Fatalf("two tunnels to one Server created %d connections, want 1", got)
	}
	if got := childCount("owner.s.edge-b"); got != 1 {
		t.Fatalf("second Server connection count = %d, want 1", got)
	}

	pulled, err := masterclient.RPCPullConfig(app.NewContext(context.Background(), application), &pb.PullClientConfigReq{
		Base: &pb.ClientBase{ClientId: "owner.c.mac", ClientSecret: "client-secret"},
	})
	if err != nil {
		t.Fatal(err)
	}
	children := append([]string(nil), pulled.GetClient().GetClientIds()...)
	sort.Strings(children)
	if len(children) != 2 {
		t.Fatalf("Client discovered %d managed FRPC connections, want 2: %#v", len(children), children)
	}

	listRecorder := httptest.NewRecorder()
	router.ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/api/v2/tunnels", nil))
	if listRecorder.Code != http.StatusOK || !strings.Contains(listRecorder.Body.String(), `"clientId":"owner.c.mac"`) || strings.Contains(listRecorder.Body.String(), `"nodeId"`) {
		t.Fatalf("list tunnels status = %d, body = %s", listRecorder.Code, listRecorder.Body.String())
	}

	remove("ssh", "owner.s.edge-a")
	if got := childCount("owner.s.edge-a"); got != 1 {
		t.Fatalf("shared connection was removed with one tunnel remaining: %d", got)
	}
	remove("web", "owner.s.edge-a")
	if got := childCount("owner.s.edge-a"); got != 0 {
		t.Fatalf("unused first Server connection remains: %d", got)
	}
	remove("dns", "owner.s.edge-b")
	if got := childCount("owner.s.edge-b"); got != 0 {
		t.Fatalf("unused second Server connection remains: %d", got)
	}
}

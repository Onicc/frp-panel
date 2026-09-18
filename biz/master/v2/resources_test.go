package v2

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func resourceTestApp(t *testing.T) (app.Application, *models.UserEntity) {
	t.Helper()
	a := app.NewApp()
	cfg := conf.DefaultConfig()
	cfg.App.EnableRegister = true
	cfg.App.CookieSecure = false
	a.SetConfig(cfg)
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "resources.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	manager := models.NewDBManager(defs.DBTypeSQLite3)
	manager.SetDB(defs.DBTypeSQLite3, defs.DBRoleDefault, db)
	manager.Init()
	a.SetDBManager(manager)
	user := &models.UserEntity{UserID: 1, TenantID: 1, UserName: "owner", Email: "owner@example.test", Role: defs.UserRole_Owner, Status: models.STATUS_NORMAL, Token: "frp-token"}
	if err := db.Create(&models.User{UserEntity: user}).Error; err != nil {
		t.Fatal(err)
	}
	return a, user
}

func resourceRouter(a app.Application, user *models.UserEntity) *gin.Engine {
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(defs.UserInfoKey, user); c.Next() })
	r.POST("/clients", createClient(a))
	r.GET("/clients", listClients(a))
	r.PATCH("/clients/:id", patchClient(a))
	r.DELETE("/clients/:id", deleteClient(a))
	r.POST("/clients/:id/enrollment", rotateClientEnrollment(a))
	r.GET("/topology", topology(a))
	r.POST("/servers", createServer(a))
	r.PATCH("/servers/:id", patchServer(a))
	r.DELETE("/servers/:id", deleteServer(a))
	r.POST("/tunnels", createTunnelResource(a))
	r.PATCH("/tunnels/:id", patchTunnel(a))
	r.DELETE("/tunnels/:id", deleteTunnelResource(a))
	return r
}

func TestTunnelUpdateChangesServerAndProtectsRemotePort(t *testing.T) {
	a, user := resourceTestApp(t)
	r := resourceRouter(a, user)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(rec, req)
		return rec
	}
	if rec := request(http.MethodPost, "/servers", `{"serverId":"edge-a","address":"a.example.test","bindPort":7100}`); rec.Code != http.StatusCreated {
		t.Fatalf("create first server = %d %s", rec.Code, rec.Body.String())
	}
	if rec := request(http.MethodPost, "/servers", `{"serverId":"edge-b","address":"b.example.test","bindPort":7200}`); rec.Code != http.StatusCreated {
		t.Fatalf("create second server = %d %s", rec.Code, rec.Body.String())
	}
	if rec := request(http.MethodPost, "/clients", `{"clientId":"mac"}`); rec.Code != http.StatusCreated {
		t.Fatalf("create client = %d %s", rec.Code, rec.Body.String())
	}
	var first struct {
		Tunnel tunnelResource `json:"tunnel"`
	}
	created := request(http.MethodPost, "/tunnels", `{"name":"ssh-a","clientId":"mac","serverId":"edge-a","type":"tcp","localPort":22,"remotePort":6022}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create first tunnel = %d %s", created.Code, created.Body.String())
	}
	if err := json.Unmarshal(created.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	conflict := request(http.MethodPost, "/tunnels", `{"name":"ssh-b","clientId":"mac","serverId":"edge-a","type":"tcp","localPort":23,"remotePort":6022}`)
	if conflict.Code != http.StatusConflict || !strings.Contains(conflict.Body.String(), "Remote port already in use") {
		t.Fatalf("remote port conflict = %d %s", conflict.Code, conflict.Body.String())
	}
	reserved := request(http.MethodPost, "/tunnels", `{"name":"ssh-reserved","clientId":"mac","serverId":"edge-a","type":"tcp","localPort":25,"remotePort":8999}`)
	if reserved.Code != http.StatusConflict || !strings.Contains(reserved.Body.String(), "SERVER_API_PORT") {
		t.Fatalf("reserved server API port = %d %s", reserved.Code, reserved.Body.String())
	}
	movable := request(http.MethodPost, "/tunnels", `{"name":"ssh-c","clientId":"mac","serverId":"edge-a","type":"tcp","localPort":24,"remotePort":6023}`)
	if movable.Code != http.StatusCreated {
		t.Fatalf("create movable tunnel = %d %s", movable.Code, movable.Body.String())
	}
	if err := json.Unmarshal(movable.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	moved := request(http.MethodPatch, "/tunnels/"+first.Tunnel.ID, `{"serverId":"edge-b"}`)
	if moved.Code != http.StatusOK {
		t.Fatalf("move tunnel = %d %s", moved.Code, moved.Body.String())
	}
	var movedPayload struct {
		Tunnel tunnelResource `json:"tunnel"`
	}
	if err := json.Unmarshal(moved.Body.Bytes(), &movedPayload); err != nil {
		t.Fatal(err)
	}
	if movedPayload.Tunnel.ServerID != "owner.s.edge-b" {
		t.Fatalf("moved tunnel server = %q", movedPayload.Tunnel.ServerID)
	}
	var stored models.ProxyConfig
	if err := a.GetDBManager().GetDefaultDB().Where("public_id = ?", first.Tunnel.ID).First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored.ServerID != "owner.s.edge-b" {
		t.Fatalf("stored tunnel server = %q", stored.ServerID)
	}
}

func TestServerAPIPortIsConfigurableAndCannotMatchBindPort(t *testing.T) {
	a, user := resourceTestApp(t)
	r := resourceRouter(a, user)
	post := func(body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/servers", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(rec, req)
		return rec
	}

	conflict := post(`{"serverId":"same-port","address":"edge.example.test","bindPort":7001,"serverApiPort":7001}`)
	if conflict.Code != http.StatusBadRequest || !strings.Contains(conflict.Body.String(), "bindPort and SERVER_API_PORT must be different") {
		t.Fatalf("same port status = %d %s", conflict.Code, conflict.Body.String())
	}

	created := post(`{"serverId":"custom-api","address":"edge.example.test","bindPort":7001,"serverApiPort":8998}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("custom API port status = %d %s", created.Code, created.Body.String())
	}
	var payload struct {
		Server     serverResource    `json:"server"`
		Enrollment enrollmentPayload `json:"enrollment"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Server.ServerAPIPort != 8998 || strings.Contains(payload.Enrollment.ComposeYAML, "\t") || !strings.Contains(payload.Enrollment.ComposeYAML, "SERVER_API_PORT: \"8998\"") {
		t.Fatalf("server API port was not propagated: %#v", payload)
	}
}

func TestTopologyIncludesTunnelEdgesAndPublicEndpointLocations(t *testing.T) {
	a, user := resourceTestApp(t)
	r := resourceRouter(a, user)
	post := func(path, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(rec, req)
		return rec
	}
	if rec := post("/clients", `{"clientId":"mac"}`); rec.Code != http.StatusCreated {
		t.Fatalf("create client = %d %s", rec.Code, rec.Body.String())
	}
	if rec := post("/servers", `{"serverId":"edge","address":"edge.example.test","bindPort":7100}`); rec.Code != http.StatusCreated {
		t.Fatalf("create server = %d %s", rec.Code, rec.Body.String())
	}
	if rec := post("/tunnels", `{"name":"ssh","clientId":"mac","serverId":"edge","type":"tcp","localPort":22,"remotePort":6022}`); rec.Code != http.StatusCreated {
		t.Fatalf("create tunnel = %d %s", rec.Code, rec.Body.String())
	}
	db := a.GetDBManager().GetDefaultDB()
	if err := db.Model(&models.Client{}).Where("client_id = ?", "owner.c.mac").Update("last_seen_ip", "8.8.8.8").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&models.Server{}).Where("server_id = ?", "owner.s.edge").Update("last_seen_ip", "1.1.1.1").Error; err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/topology", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("topology = %d %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Nodes []topologyNode `json:"nodes"`
		Links []topologyLink `json:"links"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Nodes) != 2 || len(payload.Links) != 1 {
		t.Fatalf("topology shape = %#v", payload)
	}
	locations := map[string]string{}
	for _, node := range payload.Nodes {
		locations[node.ID] = node.LocationIP
	}
	if locations["owner.c.mac"] != "8.8.8.8" || locations["owner.s.edge"] != "1.1.1.1" {
		t.Fatalf("topology locations = %#v", locations)
	}
	if payload.Links[0].SourceClientID != "owner.c.mac" || payload.Links[0].TargetServerID != "owner.s.edge" {
		t.Fatalf("topology edge = %#v", payload.Links[0])
	}
}

func TestResourceLifecycleAndDependencyProtection(t *testing.T) {
	a, user := resourceTestApp(t)
	r := resourceRouter(a, user)
	post := func(path, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(rec, req)
		return rec
	}
	clientRec := post("/clients", `{"clientId":"mac","comment":"office"}`)
	if clientRec.Code != http.StatusCreated {
		t.Fatalf("create client = %d %s", clientRec.Code, clientRec.Body.String())
	}
	var clientPayload struct {
		Client clientResource `json:"client"`
	}
	if err := json.Unmarshal(clientRec.Body.Bytes(), &clientPayload); err != nil {
		t.Fatal(err)
	}
	if clientPayload.Client.ConfigurationState != "unconfigured" || clientPayload.Client.Status != "pending" {
		t.Fatalf("new client state = %#v", clientPayload.Client)
	}
	serverRec := post("/servers", `{"serverId":"edge","address":"edge.example.test","bindPort":7100}`)
	if serverRec.Code != http.StatusCreated {
		t.Fatalf("create server = %d %s", serverRec.Code, serverRec.Body.String())
	}
	if strings.Count(serverRec.Body.String(), "PUBLIC_URL:") != 1 || !strings.Contains(serverRec.Body.String(), "SERVER_ENROLLMENT_TOKEN:") {
		t.Fatalf("server enrollment compose is incomplete: %s", serverRec.Body.String())
	}
	tunnelRec := post("/tunnels", `{"name":"ssh","clientId":"owner.c.mac","serverId":"owner.s.edge","type":"tcp","localHost":"127.0.0.1","localPort":22,"remotePort":6022}`)
	if tunnelRec.Code != http.StatusCreated {
		t.Fatalf("create tunnel = %d %s", tunnelRec.Code, tunnelRec.Body.String())
	}
	var tunnelPayload struct {
		Tunnel tunnelResource `json:"tunnel"`
	}
	if err := json.Unmarshal(tunnelRec.Body.Bytes(), &tunnelPayload); err != nil {
		t.Fatal(err)
	}
	if tunnelPayload.Tunnel.Status != "pending" {
		t.Fatalf("new tunnel state = %#v", tunnelPayload.Tunnel)
	}
	deleteClient := httptest.NewRecorder()
	r.ServeHTTP(deleteClient, httptest.NewRequest(http.MethodDelete, "/clients/owner.c.mac", nil))
	if deleteClient.Code != http.StatusConflict || !bytes.Contains(deleteClient.Body.Bytes(), []byte(`"tunnels"`)) {
		t.Fatalf("dependent client delete = %d %s", deleteClient.Code, deleteClient.Body.String())
	}
	if tunnelPayload.Tunnel.ID == "" {
		t.Fatal("tunnel id was not assigned")
	}
	delTunnel := httptest.NewRecorder()
	r.ServeHTTP(delTunnel, httptest.NewRequest(http.MethodDelete, "/tunnels/"+tunnelPayload.Tunnel.ID, nil))
	if delTunnel.Code != http.StatusNoContent {
		t.Fatalf("delete tunnel = %d %s", delTunnel.Code, delTunnel.Body.String())
	}
	delClient := httptest.NewRecorder()
	r.ServeHTTP(delClient, httptest.NewRequest(http.MethodDelete, "/clients/owner.c.mac", nil))
	if delClient.Code != http.StatusNoContent {
		t.Fatalf("delete client = %d %s", delClient.Code, delClient.Body.String())
	}
}

func TestConfiguredCredentialRotationRequiresAcknowledgement(t *testing.T) {
	a, user := resourceTestApp(t)
	r := resourceRouter(a, user)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/clients", bytes.NewBufferString(`{"clientId":"mac"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	var payload struct {
		Client clientResource `json:"client"`
	}
	json.Unmarshal(rec.Body.Bytes(), &payload)
	now := time.Now().UTC()
	a.GetDBManager().GetDefaultDB().Model(&models.Client{}).Where("client_id = ?", payload.Client.ID).Update("enrolled_at", now)
	rotation := httptest.NewRecorder()
	r.ServeHTTP(rotation, httptest.NewRequest(http.MethodPost, "/clients/owner.c.mac/enrollment", bytes.NewBufferString(`{}`)))
	if rotation.Code != http.StatusConflict {
		t.Fatalf("rotation without acknowledgement = %d %s", rotation.Code, rotation.Body.String())
	}
}

func TestEnrollmentIncludesSafeCrossPlatformInstallCommands(t *testing.T) {
	a := app.NewApp()
	cfg := conf.DefaultConfig()
	cfg.PublicURL = "https://panel.example.test"
	cfg.Client.RPCUrl = "wss://panel.example.test"
	cfg.App.AgentInstallURL = "https://raw.example.test/frp-panel"
	a.SetConfig(cfg)
	payload := makeEnrollment(a, "owner.c.mac", "client", 0, "token-without-shell-breakout", time.Now().UTC().Add(time.Minute))
	if payload.InstallCommand == "" || payload.InstallCommands["linux"] != payload.InstallCommand {
		t.Fatalf("linux compatibility command missing: %#v", payload)
	}
	for _, platform := range []string{"linux", "darwin", "windows"} {
		command := payload.InstallCommands[platform]
		if command == "" || !strings.Contains(command, "https://raw.example.test/frp-panel/install") || strings.Contains(command, "https://panel.example.test/install.sh") {
			t.Fatalf("invalid %s install command: %s", platform, command)
		}
	}
	if !strings.Contains(payload.InstallCommands["windows"], "install.ps1") || !strings.Contains(payload.InstallCommands["windows"], "--client-id") {
		t.Fatalf("windows command is incomplete: %s", payload.InstallCommands["windows"])
	}
}

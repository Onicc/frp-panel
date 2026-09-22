package dao

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSameNamedTunnelsHaveSeparateTrafficStats(t *testing.T) {
	application := app.NewApp()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "stats.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	manager := models.NewDBManager(defs.DBTypeSQLite3)
	manager.SetDB(defs.DBTypeSQLite3, defs.DBRoleDefault, db)
	manager.Init()
	application.SetDBManager(manager)
	if err := db.Create(&models.User{UserEntity: &models.UserEntity{UserID: 1, TenantID: 1, UserName: "owner"}}).Error; err != nil {
		t.Fatal(err)
	}
	for _, clientID := range []string{"owner.c.one", "owner.c.two"} {
		row := &models.ProxyConfig{ProxyConfigEntity: &models.ProxyConfigEntity{
			PublicID: clientID, UserID: 1, TenantID: 1, ServerID: "owner.s.edge",
			ClientID: clientID, OriginClientID: clientID, Name: "ssh", ManagedBy: "tunnel",
		}}
		if err := db.Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	count := func(value int64) *int64 { return &value }
	name := func(value string) *string { return &value }
	inputs := []*pb.ProxyInfo{
		{Name: name(utils.FRPClientUser("owner", "owner.c.one") + ".ssh"), Type: name("tcp"), TodayTrafficIn: count(100)},
		{Name: name(utils.FRPClientUser("owner", "owner.c.two") + ".ssh"), Type: name("tcp"), TodayTrafficIn: count(200)},
	}
	mutation := NewMutation(app.NewContext(context.Background(), application))
	server := &models.ServerEntity{ServerID: "owner.s.edge", UserID: 1, TenantID: 1}
	if err := mutation.AdminUpdateProxyStats(server, inputs); err != nil {
		t.Fatal(err)
	}
	var stats []models.ProxyStats
	if err := db.Where("server_id = ?", server.ServerID).Find(&stats).Error; err != nil {
		t.Fatal(err)
	}
	if len(stats) != 2 {
		t.Fatalf("got %d stats rows for two same-named Tunnels", len(stats))
	}
	byClient := map[string]int64{}
	for _, row := range stats {
		byClient[row.ClientID] = row.TodayTrafficIn
	}
	if byClient["owner.c.one"] != 100 || byClient["owner.c.two"] != 200 {
		t.Fatalf("same-named Tunnel traffic was mixed: %#v", byClient)
	}
}

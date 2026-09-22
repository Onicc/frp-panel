package versions

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/rpc"
	"google.golang.org/protobuf/proto"
)

func RefreshAll(application app.Application) {
	manager := application.GetClientsManager()
	if manager == nil {
		return
	}
	db := application.GetDBManager().GetDefaultDB()
	var clients []models.Client
	var servers []models.Server
	if db.Where("is_shadow = ? AND ephemeral = ?", false, false).Find(&clients).Error != nil {
		return
	}
	if db.Find(&servers).Error != nil {
		return
	}
	limit := make(chan struct{}, 8)
	var wg sync.WaitGroup
	for _, entry := range clients {
		if manager.Get(entry.ClientID) == nil {
			continue
		}
		wg.Add(1)
		limit <- struct{}{}
		go func(id string) {
			defer wg.Done()
			defer func() { <-limit }()
			_, _ = RefreshOne(application, "client", id)
		}(entry.ClientID)
	}
	for _, entry := range servers {
		if manager.Get(entry.ServerID) == nil {
			continue
		}
		wg.Add(1)
		limit <- struct{}{}
		go func(id string) {
			defer wg.Done()
			defer func() { <-limit }()
			_, _ = RefreshOne(application, "server", id)
		}(entry.ServerID)
	}
	wg.Wait()
}

func RefreshOne(application app.Application, kind, id string) (*conf.VersionInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	response, err := rpc.CallClientWithContext(app.NewContext(ctx, application), ctx, id, pb.Event_EVENT_PING, &pb.CommonRequest{})
	if err != nil {
		return nil, err
	}
	if response.GetEvent() != pb.Event_EVENT_PONG {
		return nil, context.DeadlineExceeded
	}
	var reported pb.ClientVersion
	if err := proto.Unmarshal(response.GetData(), &reported); err != nil {
		return nil, err
	}
	if reported.GetGitVersion() == "" {
		return nil, context.DeadlineExceeded
	}
	version := &conf.VersionInfo{GitVersion: reported.GetGitVersion(), GitCommit: reported.GetGitCommit(), GitBranch: reported.GetGitBranch(), BuildDate: reported.GetBuildDate(), GoVersion: reported.GetGoVersion(), Compiler: reported.GetCompiler(), Platform: reported.GetPlatform()}
	raw, err := json.Marshal(version)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	db := application.GetDBManager().GetDefaultDB()
	if kind == "client" {
		err = db.Model(&models.Client{}).Where("client_id = ?", id).Updates(map[string]any{"last_version": string(raw), "version_at": now}).Error
	} else {
		err = db.Model(&models.Server{}).Where("server_id = ?", id).Updates(map[string]any{"last_version": string(raw), "version_at": now}).Error
	}
	return version, err
}

func Decode(value string) *conf.VersionInfo {
	if value == "" {
		return nil
	}
	var version conf.VersionInfo
	if json.Unmarshal([]byte(value), &version) != nil {
		return nil
	}
	return &version
}

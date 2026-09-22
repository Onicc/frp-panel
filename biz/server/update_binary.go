package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/internal/containerupdate"
	"github.com/Onicc/frp-panel/internal/release"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/google/uuid"
)

var serverUpdateMu sync.Mutex

// HandleUpdateRequest uses the existing authenticated control RPC. Only the
// official resolved release is accepted; legacy fields never become paths or
// download URLs. Empty Version is a status query for one operation.
func HandleUpdateRequest(_ *app.Context, req *pb.UpgradeFrppRequest) (*pb.UpgradeFrppResponse, error) {
	operationID := req.GetWorkdir()
	if _, err := uuid.Parse(operationID); err != nil {
		return nil, fmt.Errorf("invalid update operation ID")
	}
	if req.GetVersion() == "" {
		status, err := containerupdate.ReadStatus()
		if err != nil {
			return nil, err
		}
		if status.OperationID != operationID {
			return nil, fmt.Errorf("update operation not found")
		}
		return &pb.UpgradeFrppResponse{Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: status.State + "|" + status.Error}}, nil
	}
	if !containerupdate.Enabled() {
		return nil, fmt.Errorf("server image must be redeployed before in-panel updates can run")
	}
	if req.GetDownloadUrl() != "" || req.GetGithubProxy() != "" || req.GetHttpProxy() != "" || req.GetServiceName() != "" || req.GetServiceArgs() != nil {
		return nil, fmt.Errorf("custom update parameters are not accepted")
	}
	version := conf.GetVersion()
	if release.Channel(version.GitVersion) == "unknown" {
		return nil, fmt.Errorf("unknown Server update channel")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	snapshot, rel := release.Default.Check(ctx, *version, true)
	cancel()
	if snapshot.Available == nil || !*snapshot.Available || rel.TagName != req.GetVersion() || rel.Commit != req.GetTargetPath() {
		return nil, fmt.Errorf("requested release is unavailable or does not match this Server")
	}
	serverUpdateMu.Lock()
	defer serverUpdateMu.Unlock()
	if status, err := containerupdate.ReadStatus(); err == nil {
		if status.OperationID == operationID && status.State != "failed" && status.State != "rolled_back" {
			return &pb.UpgradeFrppResponse{Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "accepted"}}, nil
		}
		if status.OperationID != operationID && (status.State == "queued" || status.State == "downloading" || status.State == "staged") {
			return nil, fmt.Errorf("another server update is already in progress")
		}
	}
	if err := containerupdate.WriteStatus(containerupdate.Status{OperationID: operationID, Commit: rel.Commit, Version: rel.TagName, State: "queued"}); err != nil {
		return nil, err
	}
	go func() {
		work, stop := context.WithTimeout(context.Background(), 12*time.Minute)
		defer stop()
		if err := containerupdate.Stage(work, rel, operationID); err != nil {
			_ = containerupdate.WriteStatus(containerupdate.Status{OperationID: operationID, Commit: rel.Commit, Version: rel.TagName, State: "failed", Error: err.Error()})
			return
		}
		containerupdate.RestartAfterResponse()
	}()
	return &pb.UpgradeFrppResponse{Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "accepted"}}, nil
}

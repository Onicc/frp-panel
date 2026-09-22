package v2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Onicc/frp-panel/biz/master/versions"
	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/internal/containerupdate"
	"github.com/Onicc/frp-panel/internal/release"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/rpc"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

var updateMu sync.Mutex
var monitoring sync.Map

var runningUpdateStates = []string{"queued", "downloading", "staged", "restarting", "verifying"}

func ownerOnly(c *gin.Context) bool {
	user, ok := currentUser(c)
	if !ok {
		AbortProblem(c, 401, "Unauthorized", "sign in first")
		return false
	}
	if user.GetRole() != defs.UserRole_Owner {
		AbortProblem(c, 403, "Forbidden", "only the Owner can update Master")
		return false
	}
	return true
}

func adminOnly(c *gin.Context) bool {
	user, ok := currentUser(c)
	if !ok {
		AbortProblem(c, 401, "Unauthorized", "sign in first")
		return false
	}
	if user.GetRole() != defs.UserRole_Owner && user.GetRole() != defs.UserRole_Admin {
		AbortProblem(c, 403, "Forbidden", "an administrator is required")
		return false
	}
	return true
}

func releaseStatus(application app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "sign in first")
			return
		}
		channel := c.Query("channel")
		if channel != "" && channel != "edge" && channel != "stable" {
			AbortProblem(c, 400, "Invalid channel", "use stable")
			return
		}
		var snapshot release.Snapshot
		if channel == "edge" {
			snapshot = release.Snapshot{Channel: "legacy", CheckedAt: time.Now().UTC(), Error: "edge builds require a one-time manual migration to a stable release"}
		} else if channel == "" || channel == release.Channel(conf.GetVersion().GitVersion) {
			snapshot, _ = release.Default.Check(c.Request.Context(), *conf.GetVersion(), c.Query("force") == "true")
		} else {
			snapshot = release.Snapshot{Channel: channel, CheckedAt: time.Now().UTC()}
			if rel, err := release.Default.Latest(c.Request.Context(), channel, c.Query("force") == "true"); err == nil {
				snapshot.LatestVersion, snapshot.LatestCommit, snapshot.ReleaseURL, snapshot.PublishedAt = rel.TagName, rel.Commit, rel.HTMLURL, rel.PublishedAt
			} else {
				snapshot.Error = err.Error()
			}
		}
		result := gin.H{"release": snapshot, "supported": containerupdate.Enabled()}
		if user.GetRole() == defs.UserRole_Owner {
			if op := latestOperation(application.GetDBManager().GetDefaultDB(), "master", "master"); op != nil {
				reconcileMasterOperation(application, op)
				result["operation"] = op
			}
		}
		c.JSON(200, result)
	}
}

func startMasterUpdate(application app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !ownerOnly(c) {
			return
		}
		if release.Channel(conf.GetVersion().GitVersion) == "legacy" {
			AbortProblem(c, 409, "Manual migration required", "redeploy the Master with a vX.X.X image before using web updates")
			return
		}
		if !containerupdate.Enabled() {
			AbortProblem(c, 409, "Image update required", "redeploy the Master image once before using web updates")
			return
		}
		snapshot, rel := release.Default.Check(c.Request.Context(), *conf.GetVersion(), true)
		if snapshot.Available == nil {
			AbortProblem(c, 503, "Version unavailable", snapshot.Error)
			return
		}
		if !*snapshot.Available {
			AbortProblem(c, 409, "Already up to date", "no newer release is available in this channel")
			return
		}
		user, _ := currentUser(c)
		op, err := createUpdateOperation(application, "master", "master", user.GetTenantID(), user.GetUserID(), "manual", rel)
		if err != nil {
			AbortProblem(c, 409, "Update in progress", err.Error())
			return
		}
		go performMasterUpdate(application, op, rel)
		c.JSON(http.StatusAccepted, gin.H{"operationId": op.ID})
	}
}

func performMasterUpdate(application app.Application, op models.UpdateOperation, rel release.Release) {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	setUpdateState(application, op.ID, "downloading", "")
	if err := containerupdate.Stage(ctx, rel, op.ID); err != nil {
		setUpdateState(application, op.ID, "failed", err.Error())
		_ = containerupdate.WriteStatus(containerupdate.Status{OperationID: op.ID, Commit: rel.Commit, Version: rel.TagName, State: "failed", Error: err.Error()})
		return
	}
	setUpdateState(application, op.ID, "restarting", "")
	containerupdate.RestartAfterResponse()
}

func updateOperation(application app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "sign in first")
			return
		}
		id := c.Param("id")
		if _, err := uuid.Parse(id); err != nil {
			AbortProblem(c, 400, "Invalid operation", "operation ID must be a UUID")
			return
		}
		var op models.UpdateOperation
		if application.GetDBManager().GetDefaultDB().Where("id = ?", id).First(&op).Error != nil {
			AbortProblem(c, 404, "Operation not found", "update operation does not exist")
			return
		}
		if op.Kind == "master" && user.GetRole() != defs.UserRole_Owner {
			AbortProblem(c, 403, "Forbidden", "Owner access required")
			return
		}
		if op.Kind == "server" && (op.UserID != user.GetUserID() || op.TenantID != user.GetTenantID()) {
			AbortProblem(c, 404, "Operation not found", "update operation does not exist")
			return
		}
		if op.Kind == "master" {
			reconcileMasterOperation(application, &op)
		}
		c.JSON(200, gin.H{"operation": op})
	}
}

func reconcileMasterOperation(application app.Application, op *models.UpdateOperation) {
	if finished(op.State) {
		return
	}
	if status, err := containerupdate.ReadStatus(); err == nil && status.OperationID == op.ID && finished(status.State) {
		setUpdateState(application, op.ID, status.State, status.Error)
		op.State, op.Error = status.State, status.Error
	}
}

func serverUpdatePolicy(application app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !adminOnly(c) {
			return
		}
		user, _ := currentUser(c)
		id := scopedID(user.GetUserName(), "s", c.Param("id"))
		var req struct {
			Enabled bool   `json:"enabled"`
			Zone    string `json:"timeZone"`
			Start   string `json:"windowStart"`
			End     string `json:"windowEnd"`
		}
		if c.ShouldBindJSON(&req) != nil {
			AbortProblem(c, 400, "Invalid update policy", "send a valid JSON object")
			return
		}
		if req.Zone == "" {
			req.Zone = "UTC"
		}
		if req.Start == "" {
			req.Start = "03:00"
		}
		if req.End == "" {
			req.End = "04:00"
		}
		if _, err := time.LoadLocation(req.Zone); err != nil {
			AbortProblem(c, 422, "Invalid timezone", "use an IANA timezone")
			return
		}
		start, okStart := timeOfDay(req.Start)
		end, okEnd := timeOfDay(req.End)
		if !okStart || !okEnd || start >= end {
			AbortProblem(c, 422, "Invalid window", "window must be a same-day HH:MM range")
			return
		}
		db := application.GetDBManager().GetDefaultDB()
		result := db.Model(&models.Server{}).Where("server_id = ? AND user_id = ? AND tenant_id = ?", id, user.GetUserID(), user.GetTenantID()).Updates(map[string]any{"auto_update": req.Enabled, "update_zone": req.Zone, "update_start": req.Start, "update_end": req.End})
		if result.Error != nil {
			AbortProblem(c, 500, "Update policy unavailable", "could not save policy")
			return
		}
		if result.RowsAffected == 0 {
			AbortProblem(c, 404, "Server not found", "Server does not exist")
			return
		}
		c.JSON(200, gin.H{"enabled": req.Enabled, "timeZone": req.Zone, "windowStart": req.Start, "windowEnd": req.End})
	}
}

func startServerUpdateHandler(application app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !adminOnly(c) {
			return
		}
		user, _ := currentUser(c)
		id := scopedID(user.GetUserName(), "s", c.Param("id"))
		var server models.Server
		if application.GetDBManager().GetDefaultDB().Where("server_id = ? AND user_id = ? AND tenant_id = ?", id, user.GetUserID(), user.GetTenantID()).First(&server).Error != nil {
			AbortProblem(c, 404, "Server not found", "Server does not exist")
			return
		}
		op, err := startServerUpdate(application, server, "manual")
		if err != nil {
			AbortProblem(c, 409, "Server update unavailable", err.Error())
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"operationId": op.ID})
	}
}

func startServerUpdate(application app.Application, server models.Server, mode string) (models.UpdateOperation, error) {
	id := server.ServerID
	if application.GetClientsManager().Get(id) == nil {
		return models.UpdateOperation{}, errors.New("server is offline")
	}
	version := versions.Decode(server.LastVersion)
	if version == nil || time.Since(valueTime(server.VersionAt)) > 2*time.Minute {
		var err error
		version, err = versions.RefreshOne(application, "server", id)
		if err != nil {
			return models.UpdateOperation{}, fmt.Errorf("server version unavailable: %w", err)
		}
	}
	if release.Channel(version.GitVersion) == "legacy" {
		return models.UpdateOperation{}, errors.New("redeploy this Server with a vX.X.X image before using panel updates")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	snapshot, rel := release.Default.Check(ctx, *version, false)
	cancel()
	if snapshot.Available == nil {
		return models.UpdateOperation{}, fmt.Errorf("release comparison unavailable: %s", snapshot.Error)
	}
	if !*snapshot.Available {
		return models.UpdateOperation{}, errors.New("server is already up to date")
	}
	op, err := createUpdateOperation(application, "server", id, server.TenantID, server.UserID, mode, rel)
	if err != nil {
		return models.UpdateOperation{}, err
	}
	request := &pb.UpgradeFrppRequest{Version: &rel.TagName, Workdir: &op.ID, TargetPath: &rel.Commit}
	callCtx, callCancel := context.WithTimeout(context.Background(), 35*time.Second)
	defer callCancel()
	response, err := rpc.CallClientWithContext(app.NewContext(callCtx, application), callCtx, id, pb.Event_EVENT_UPGRADE_FRPP, request)
	if err != nil {
		setUpdateState(application, op.ID, "failed", err.Error())
		return models.UpdateOperation{}, err
	}
	var decoded pb.UpgradeFrppResponse
	if err = proto.Unmarshal(response.GetData(), &decoded); err != nil {
		setUpdateState(application, op.ID, "failed", err.Error())
		return models.UpdateOperation{}, err
	}
	if decoded.GetStatus().GetCode() != pb.RespCode_RESP_CODE_SUCCESS {
		err = fmt.Errorf("server refused update: %s", decoded.GetStatus().GetMessage())
		setUpdateState(application, op.ID, "failed", err.Error())
		return models.UpdateOperation{}, err
	}
	setUpdateState(application, op.ID, "downloading", "")
	monitorServerUpdate(application, op)
	return op, nil
}

func monitorServerUpdate(application app.Application, op models.UpdateOperation) {
	if _, loaded := monitoring.LoadOrStore(op.ID, struct{}{}); loaded {
		return
	}
	go func() {
		defer monitoring.Delete(op.ID)
		ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
		defer cancel()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				setUpdateState(application, op.ID, "failed", "Server did not confirm the target version within 12 minutes")
				return
			case <-ticker.C:
				version, err := versions.RefreshOne(application, "server", op.TargetID)
				if err == nil && strings.EqualFold(version.GitCommit, op.TargetCommit) {
					setUpdateState(application, op.ID, "succeeded", "")
					return
				}
				if application.GetClientsManager().Get(op.TargetID) == nil {
					continue
				}
				query := &pb.UpgradeFrppRequest{Workdir: &op.ID}
				callCtx, stop := context.WithTimeout(ctx, 3*time.Second)
				response, queryErr := rpc.CallClientWithContext(app.NewContext(callCtx, application), callCtx, op.TargetID, pb.Event_EVENT_UPGRADE_FRPP, query)
				stop()
				if queryErr != nil {
					continue
				}
				var result pb.UpgradeFrppResponse
				if proto.Unmarshal(response.GetData(), &result) != nil {
					continue
				}
				message := result.GetStatus().GetMessage()
				if strings.HasPrefix(message, "failed|") || strings.HasPrefix(message, "rolled_back|") {
					parts := strings.SplitN(message, "|", 2)
					setUpdateState(application, op.ID, parts[0], parts[1])
					return
				}
				if strings.HasPrefix(message, "staged|") {
					setUpdateState(application, op.ID, "staged", "")
				}
				if strings.HasPrefix(message, "succeeded|") {
					setUpdateState(application, op.ID, "verifying", "")
				}
			}
		}
	}()
}

func ResumeUpdateOperations(application app.Application) {
	db := application.GetDBManager().GetDefaultDB()
	var ops []models.UpdateOperation
	if db.Where("kind = ? AND state IN ?", "master", runningUpdateStates).Find(&ops).Error == nil {
		for _, op := range ops {
			state, reason := masterRestartDecision(op.ID)
			if state != "" {
				setUpdateState(application, op.ID, state, reason)
			}
		}
	}
	ops = nil
	if db.Where("kind = ? AND state IN ?", "server", runningUpdateStates).Find(&ops).Error != nil {
		return
	}
	for _, op := range ops {
		monitorServerUpdate(application, op)
	}
}

func masterRestartDecision(operationID string) (string, string) {
	status, err := containerupdate.ReadStatus()
	if err == nil && status.OperationID == operationID {
		if finished(status.State) {
			return status.State, status.Error
		}
		if status.State == "staged" {
			// Master starts before the launcher finishes its readiness probe.
			// Keep the operation active until promotion or rollback is recorded.
			if manifest, pendingErr := containerupdate.ReadManifest(containerupdate.PendingPath()); pendingErr == nil && manifest.OperationID == operationID {
				return "", ""
			}
			return "failed", "staged master update has no matching pending manifest"
		}
	}
	return "failed", "master update was interrupted before a verified restart"
}

func ScheduleServerUpdates(application app.Application) {
	ResumeUpdateOperations(application)
	db := application.GetDBManager().GetDefaultDB()
	var active int64
	if db.Model(&models.UpdateOperation{}).Where("kind = ? AND mode = ? AND state IN ?", "server", "auto", runningUpdateStates).Count(&active).Error != nil || active > 0 {
		return
	}
	var servers []models.Server
	if db.Where("auto_update = ?", true).Order("server_id ASC").Find(&servers).Error != nil {
		return
	}
	now := time.Now().UTC()
	for _, server := range servers {
		if !withinWindow(server.ServerEntity, now) || application.GetClientsManager().Get(server.ServerID) == nil {
			continue
		}
		if version := versions.Decode(server.LastVersion); version == nil || release.Channel(version.GitVersion) != "stable" {
			continue
		}
		var previous models.UpdateOperation
		if db.Where("kind = ? AND target_id = ? AND mode = ?", "server", server.ServerID, "auto").Order("started_at DESC").First(&previous).Error == nil {
			zone, _ := time.LoadLocation(server.UpdateZone)
			if zone != nil && previous.StartedAt.In(zone).Format("2006-01-02") == now.In(zone).Format("2006-01-02") {
				continue
			}
		}
		if _, err := startServerUpdate(application, server, "auto"); err == nil {
			return
		}
	}
}

func withinWindow(server *models.ServerEntity, now time.Time) bool {
	if server == nil {
		return false
	}
	zoneName := server.UpdateZone
	if zoneName == "" {
		zoneName = "UTC"
	}
	zone, err := time.LoadLocation(zoneName)
	if err != nil {
		return false
	}
	start, ok1 := timeOfDay(server.UpdateStart)
	end, ok2 := timeOfDay(server.UpdateEnd)
	if !ok1 || !ok2 || start >= end {
		return false
	}
	local := now.In(zone)
	minute := local.Hour()*60 + local.Minute()
	return minute >= start && minute < end
}

func timeOfDay(value string) (int, bool) {
	if len(value) != 5 || value[2] != ':' {
		return 0, false
	}
	hour, errH := strconv.Atoi(value[:2])
	minute, errM := strconv.Atoi(value[3:])
	if errH != nil || errM != nil || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, false
	}
	return hour*60 + minute, true
}

func valueTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}
func finished(state string) bool {
	return state == "succeeded" || state == "failed" || state == "rolled_back"
}

func createUpdateOperation(application app.Application, kind, target string, tenant, user int, mode string, rel release.Release) (models.UpdateOperation, error) {
	updateMu.Lock()
	defer updateMu.Unlock()
	db := application.GetDBManager().GetDefaultDB()
	var count int64
	if err := db.Model(&models.UpdateOperation{}).Where("kind = ? AND target_id = ? AND state IN ?", kind, target, runningUpdateStates).Count(&count).Error; err != nil {
		return models.UpdateOperation{}, err
	}
	if count > 0 {
		return models.UpdateOperation{}, errors.New("another update is already running")
	}
	op := models.UpdateOperation{ID: uuid.NewString(), Kind: kind, TargetID: target, TenantID: tenant, UserID: user, Mode: mode, TargetCommit: rel.Commit, Version: rel.TagName, State: "queued", StartedAt: time.Now().UTC()}
	if err := db.Create(&op).Error; err != nil {
		return models.UpdateOperation{}, err
	}
	return op, nil
}

func setUpdateState(application app.Application, id, state, reason string) {
	updates := map[string]any{"state": state, "error": reason, "updated_at": time.Now().UTC()}
	if finished(state) {
		now := time.Now().UTC()
		updates["completed_at"] = now
	}
	_ = application.GetDBManager().GetDefaultDB().Model(&models.UpdateOperation{}).Where("id = ?", id).Updates(updates).Error
}

func latestOperation(db *gorm.DB, kind, target string) *models.UpdateOperation {
	var op models.UpdateOperation
	if db.Where("kind = ? AND target_id = ?", kind, target).Order("started_at DESC").First(&op).Error != nil {
		return nil
	}
	return &op
}

func reportedVersion(value string) *conf.VersionInfo { return versions.Decode(value) }

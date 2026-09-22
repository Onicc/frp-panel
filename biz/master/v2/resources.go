package v2

// This file contains the resource-oriented management API. A Client is a
// physical business host, a Server is an FRPS data-plane host, and a Tunnel is
// the only user-created forwarding resource. Internal FRPC connections remain
// implementation details and are never returned as Clients.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Onicc/frp-panel/common"
	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/rpc"
	"github.com/Onicc/frp-panel/utils"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const defaultPageSize = 25

var tunnelPairLocks sync.Map
var tunnelServerLocks sync.Map

type pageResult[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

type clientResource struct {
	ID                 string            `json:"id"`
	Comment            string            `json:"comment"`
	ConfigurationState string            `json:"configurationState"`
	Status             string            `json:"status"`
	Enabled            bool              `json:"enabled"`
	LastSeenAt         *time.Time        `json:"lastSeenAt,omitempty"`
	EnrolledAt         *time.Time        `json:"enrolledAt,omitempty"`
	TunnelCount        int64             `json:"tunnelCount"`
	LocationIPOverride string            `json:"locationIpOverride,omitempty"`
	Version            *conf.VersionInfo `json:"version,omitempty"`
	VersionAt          *time.Time        `json:"versionAt,omitempty"`
}

type serverResource struct {
	ID                 string                  `json:"id"`
	Address            string                  `json:"address"`
	BindPort           int                     `json:"bindPort"`
	ServerAPIPort      int                     `json:"serverApiPort"`
	Comment            string                  `json:"comment"`
	ConfigurationState string                  `json:"configurationState"`
	Status             string                  `json:"status"`
	LastSeenAt         *time.Time              `json:"lastSeenAt,omitempty"`
	EnrolledAt         *time.Time              `json:"enrolledAt,omitempty"`
	TunnelCount        int64                   `json:"tunnelCount"`
	Version            *conf.VersionInfo       `json:"version,omitempty"`
	VersionAt          *time.Time              `json:"versionAt,omitempty"`
	AutoUpdate         bool                    `json:"autoUpdate"`
	UpdateZone         string                  `json:"updateZone"`
	UpdateStart        string                  `json:"updateStart"`
	UpdateEnd          string                  `json:"updateEnd"`
	UpdateOperation    *models.UpdateOperation `json:"updateOperation,omitempty"`
}

type topologyNode struct {
	ID                 string     `json:"id"`
	Kind               string     `json:"kind"`
	Label              string     `json:"label"`
	Comment            string     `json:"comment,omitempty"`
	Address            string     `json:"address,omitempty"`
	Status             string     `json:"status"`
	ConfigurationState string     `json:"configurationState"`
	Enabled            bool       `json:"enabled"`
	LocationIP         string     `json:"locationIp,omitempty"`
	LocationSource     string     `json:"locationSource,omitempty"`
	ObservedIP         string     `json:"observedIp,omitempty"`
	ReportedIP         string     `json:"reportedIp,omitempty"`
	ReportedAt         *time.Time `json:"reportedAt,omitempty"`
	LastSeenAt         *time.Time `json:"lastSeenAt,omitempty"`
	TunnelCount        int64      `json:"tunnelCount"`
}

type topologyLink struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	SourceClientID string `json:"sourceClientId"`
	TargetServerID string `json:"targetServerId"`
	Type           string `json:"type"`
	RemotePort     int    `json:"remotePort"`
	Enabled        bool   `json:"enabled"`
	Status         string `json:"status"`
	LastError      string `json:"lastError,omitempty"`
}

type tunnelResource struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	ClientID      string    `json:"clientId"`
	ServerID      string    `json:"serverId"`
	ServerAddress string    `json:"serverAddress,omitempty"`
	Type          string    `json:"type"`
	LocalHost     string    `json:"localHost"`
	LocalPort     int       `json:"localPort"`
	RemotePort    int       `json:"remotePort"`
	Enabled       bool      `json:"enabled"`
	Status        string    `json:"status"`
	LastError     string    `json:"lastError,omitempty"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type enrollmentPayload struct {
	ClientID        string            `json:"clientId,omitempty"`
	ServerID        string            `json:"serverId,omitempty"`
	Token           string            `json:"token"`
	ExpiresAt       time.Time         `json:"expiresAt"`
	APIURL          string            `json:"apiUrl"`
	RPCURL          string            `json:"rpcUrl"`
	InstallCommand  string            `json:"installCommand,omitempty"`
	InstallCommands map[string]string `json:"installCommands,omitempty"`
	ComposeYAML     string            `json:"composeYaml,omitempty"`
}

type clientCreateRequest struct {
	ClientID string `json:"clientId"`
	Comment  string `json:"comment"`
}
type clientPatchRequest struct {
	ClientID           *string `json:"clientId"`
	ID                 *string `json:"id"`
	Comment            *string `json:"comment"`
	Enabled            *bool   `json:"enabled"`
	LocationIPOverride *string `json:"locationIpOverride"`
}
type serverCreateRequest struct {
	ServerID      string `json:"serverId"`
	Address       string `json:"address"`
	BindPort      int    `json:"bindPort"`
	ServerAPIPort int    `json:"serverApiPort"`
	Comment       string `json:"comment"`
}
type serverPatchRequest struct {
	ServerID      *string `json:"serverId"`
	ID            *string `json:"id"`
	Address       *string `json:"address"`
	BindPort      *int    `json:"bindPort"`
	ServerAPIPort *int    `json:"serverApiPort"`
	Comment       *string `json:"comment"`
}
type rotateEnrollmentRequest struct {
	AcknowledgeDisruption bool `json:"acknowledgeDisruption"`
}
type tunnelRequest struct {
	Name       string `json:"name"`
	ClientID   string `json:"clientId"`
	ServerID   string `json:"serverId"`
	Type       string `json:"type"`
	LocalHost  string `json:"localHost"`
	LocalPort  int    `json:"localPort"`
	RemotePort int    `json:"remotePort"`
	Enabled    *bool  `json:"enabled"`
}
type tunnelPatchRequest struct {
	Name       *string `json:"name"`
	ClientID   *string `json:"clientId"`
	ServerID   *string `json:"serverId"`
	Type       *string `json:"type"`
	LocalHost  *string `json:"localHost"`
	LocalPort  *int    `json:"localPort"`
	RemotePort *int    `json:"remotePort"`
	Enabled    *bool   `json:"enabled"`
}

func currentUser(c *gin.Context) (models.UserInfo, bool) {
	u := common.GetUserInfo(c)
	return u, u != nil && u.Valid()
}

func scopedID(username, kind, value string) string {
	value = strings.TrimSpace(value)
	if strings.Count(value, ".") >= 2 {
		return value
	}
	return app.GlobalClientID(username, kind, value)
}

func resourcePage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", strconv.Itoa(defaultPageSize)))
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = defaultPageSize
	}
	if size > 100 {
		size = 100
	}
	return page, size
}

func pageSlice[T any](items []T, page, size int) []T {
	start := (page - 1) * size
	if start >= len(items) {
		return []T{}
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func isConfigured(enrolledAt *time.Time, content []byte) bool {
	// Enrollment, not merely a locally generated config, is the source of
	// truth. Migrations backfill EnrolledAt for pre-v2 rows that already had a
	// config, while newly created resources remain visibly unconfigured until
	// their one-use token is redeemed.
	return enrolledAt != nil
}

func connectorOnline(appInstance app.Application, id string) bool {
	manager := appInstance.GetClientsManager()
	return manager != nil && manager.Get(id) != nil
}

func clientStatus(appInstance app.Application, item *models.ClientEntity) string {
	if !item.Enabled || item.Stopped {
		return "disabled"
	}
	if !isConfigured(item.EnrolledAt, item.ConfigContent) {
		return "pending"
	}
	if connectorOnline(appInstance, item.ClientID) {
		// The connector is authenticated and its receive loop is alive. A
		// bounded ping is best-effort; failure is represented as error rather
		// than holding the HTTP request open indefinitely.
		return "online"
	}
	return "offline"
}

func serverStatus(appInstance app.Application, item *models.ServerEntity) string {
	if !isConfigured(item.EnrolledAt, item.ConfigContent) {
		return "pending"
	}
	if connectorOnline(appInstance, item.ServerID) {
		return "online"
	}
	return "offline"
}

func tunnelStatus(appInstance app.Application, item *models.ProxyConfig) string {
	if item.Stopped {
		return "disabled"
	}
	client := &models.Client{}
	server := &models.Server{}
	db := appInstance.GetDBManager().GetDefaultDB()
	if db.Where("client_id = ?", item.OriginClientID).First(client).Error != nil {
		return "error"
	}
	if db.Where("server_id = ?", item.ServerID).First(server).Error != nil {
		return "error"
	}
	if !isConfigured(client.EnrolledAt, client.ConfigContent) || !isConfigured(server.EnrolledAt, server.ConfigContent) {
		return "pending"
	}
	if item.LastError != "" {
		return "error"
	}
	if connectorOnline(appInstance, client.ClientID) && connectorOnline(appInstance, server.ServerID) {
		return "online"
	}
	return "offline"
}

func clientToResource(appInstance app.Application, db *gorm.DB, item *models.Client) clientResource {
	var count int64
	db.Model(&models.ProxyConfig{}).Where("tenant_id = ? AND user_id = ? AND managed_by = ? AND origin_client_id = ?", item.TenantID, item.UserID, "tunnel", item.ClientID).Count(&count)
	return clientResource{ID: item.ClientID, Comment: item.Comment, ConfigurationState: map[bool]string{true: "configured", false: "unconfigured"}[isConfigured(item.EnrolledAt, item.ConfigContent)], Status: clientStatus(appInstance, item.ClientEntity), Enabled: item.Enabled && !item.Stopped, LastSeenAt: item.LastSeenAt, EnrolledAt: item.EnrolledAt, TunnelCount: count, LocationIPOverride: item.LocationIPOverride, Version: reportedVersion(item.LastVersion), VersionAt: item.VersionAt}
}

func serverToResource(appInstance app.Application, db *gorm.DB, item *models.Server) serverResource {
	var count int64
	db.Model(&models.ProxyConfig{}).Where("tenant_id = ? AND user_id = ? AND managed_by = ? AND server_id = ?", item.TenantID, item.UserID, "tunnel", item.ServerID).Count(&count)
	serverAPIPort := item.ServerAPIPort
	if serverAPIPort == 0 {
		serverAPIPort = defs.DefaultServerAPIPort
	}
	zone, start, end := item.UpdateZone, item.UpdateStart, item.UpdateEnd
	if zone == "" {
		zone = "UTC"
	}
	if start == "" {
		start = "03:00"
	}
	if end == "" {
		end = "04:00"
	}
	return serverResource{ID: item.ServerID, Address: item.ServerIP, BindPort: item.BindPort, ServerAPIPort: serverAPIPort, Comment: item.Comment, ConfigurationState: map[bool]string{true: "configured", false: "unconfigured"}[isConfigured(item.EnrolledAt, item.ConfigContent)], Status: serverStatus(appInstance, item.ServerEntity), LastSeenAt: item.LastSeenAt, EnrolledAt: item.EnrolledAt, TunnelCount: count, Version: reportedVersion(item.LastVersion), VersionAt: item.VersionAt, AutoUpdate: item.AutoUpdate, UpdateZone: zone, UpdateStart: start, UpdateEnd: end, UpdateOperation: latestOperation(db, "server", item.ServerID)}
}

func normalizeServerPorts(bindPort, serverAPIPort int) (int, int, error) {
	if bindPort == 0 {
		bindPort = defs.DefaultFRPSBindPort
	}
	if serverAPIPort == 0 {
		serverAPIPort = defs.DefaultServerAPIPort
	}
	if bindPort < 1024 || bindPort > 65535 {
		return 0, 0, fmt.Errorf("bindPort must be between 1024 and 65535")
	}
	if serverAPIPort < 1024 || serverAPIPort > 65535 {
		return 0, 0, fmt.Errorf("SERVER_API_PORT must be between 1024 and 65535")
	}
	if bindPort == serverAPIPort {
		return 0, 0, fmt.Errorf("bindPort and SERVER_API_PORT must be different")
	}
	return bindPort, serverAPIPort, nil
}

func createClient(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		var req clientCreateRequest
		if c.ShouldBindJSON(&req) != nil || !utils.IsClientIDPermited(strings.TrimSpace(req.ClientID)) {
			AbortProblem(c, 400, "Invalid Client", "clientId must use letters, numbers, underscores, or hyphens")
			return
		}
		id := scopedID(user.GetUserName(), "c", req.ClientID)
		token, err := randomToken()
		if err != nil {
			AbortProblem(c, 500, "Client creation failed", "could not generate a secure enrollment token")
			return
		}
		expires := time.Now().UTC().Add(enrollmentLifetime)
		secret := utils.DeriveCredential(appInstance.GetConfig().App.GlobalSecret, "client-agent", token)
		db := appInstance.GetDBManager().GetDefaultDB()
		var client models.Client
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("client_id = ?", id).First(&client).Error; err == nil {
				return gorm.ErrDuplicatedKey
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			client = models.Client{ClientEntity: &models.ClientEntity{ClientID: id, TenantID: user.GetTenantID(), UserID: user.GetUserID(), Comment: strings.TrimSpace(req.Comment), ConnectSecret: utils.HashCredential(secret), Enabled: true}}
			if err := tx.Create(&client).Error; err != nil {
				return err
			}
			return tx.Create(&models.AgentEnrollment{ID: uuid.NewString(), TokenHash: utils.HashCredential(token), ClientID: id, UserID: user.GetUserID(), TenantID: user.GetTenantID(), ExpiresAt: expires}).Error
		})
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "unique") {
				AbortProblem(c, 409, "Client already exists", "choose a different Client ID")
			} else {
				AbortProblem(c, 500, "Client creation failed", "could not save the Client")
			}
			return
		}
		c.Header("Cache-Control", "no-store")
		c.JSON(201, gin.H{"client": clientToResource(appInstance, db, &client), "enrollment": makeEnrollment(appInstance, id, "client", 0, token, expires)})
	}
}

func getClient(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		id := scopedID(user.GetUserName(), "c", c.Param("id"))
		var item models.Client
		if err := appInstance.GetDBManager().GetDefaultDB().Where("client_id = ? AND user_id = ? AND tenant_id = ? AND origin_client_id = ?", id, user.GetUserID(), user.GetTenantID(), "").First(&item).Error; err != nil {
			AbortProblem(c, 404, "Client not found", "the requested Client does not exist")
			return
		}
		c.JSON(200, gin.H{"client": clientToResource(appInstance, appInstance.GetDBManager().GetDefaultDB(), &item)})
	}
}

func listClients(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		page, size := resourcePage(c)
		db := appInstance.GetDBManager().GetDefaultDB()
		q := db.Where("user_id = ? AND tenant_id = ? AND (origin_client_id IS NULL OR origin_client_id = ?)", user.GetUserID(), user.GetTenantID(), "")
		if search := strings.TrimSpace(c.Query("search")); search != "" {
			q = q.Where(db.Where("client_id LIKE ?", "%"+search+"%").Or("comment LIKE ?", "%"+search+"%"))
		}
		state := c.Query("configurationState")
		if state == "configured" {
			q = q.Where("enrolled_at IS NOT NULL")
		} else if state == "unconfigured" {
			q = q.Where("enrolled_at IS NULL")
		}
		status := c.Query("status")
		memoryStatus := status == "online" || status == "offline" || status == "error"
		if status == "pending" {
			q = q.Where("enrolled_at IS NULL AND enabled = ? AND stopped = ?", true, false)
		} else if status == "disabled" {
			q = q.Where("(enabled = ? OR stopped = ?)", false, true)
		} else if status != "" && !memoryStatus {
			q = q.Where("1 = 0")
		}
		if memoryStatus {
			var all []models.Client
			if err := q.Order("created_at DESC").Find(&all).Error; err != nil {
				AbortProblem(c, 500, "Clients unavailable", "could not list Clients")
				return
			}
			items := make([]clientResource, 0, len(all))
			for i := range all {
				r := clientToResource(appInstance, db, &all[i])
				if r.Status == status {
					items = append(items, r)
				}
			}
			c.JSON(200, pageResult[clientResource]{Items: pageSlice(items, page, size), Total: int64(len(items)), Page: page, PageSize: size})
			return
		}
		var total int64
		if err := q.Model(&models.Client{}).Count(&total).Error; err != nil {
			AbortProblem(c, 500, "Clients unavailable", "could not count Clients")
			return
		}
		var rows []models.Client
		if err := q.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
			AbortProblem(c, 500, "Clients unavailable", "could not list Clients")
			return
		}
		items := make([]clientResource, 0, len(rows))
		for i := range rows {
			r := clientToResource(appInstance, db, &rows[i])
			items = append(items, r)
		}
		c.JSON(200, pageResult[clientResource]{Items: items, Total: total, Page: page, PageSize: size})
	}
}

func patchClient(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		id := scopedID(user.GetUserName(), "c", c.Param("id"))
		var item models.Client
		db := appInstance.GetDBManager().GetDefaultDB()
		if db.Where("client_id = ? AND user_id = ? AND tenant_id = ? AND (origin_client_id IS NULL OR origin_client_id = ?)", id, user.GetUserID(), user.GetTenantID(), "").First(&item).Error != nil {
			AbortProblem(c, 404, "Client not found", "the requested Client does not exist")
			return
		}
		var req clientPatchRequest
		if c.ShouldBindJSON(&req) != nil {
			AbortProblem(c, 400, "Invalid Client update", "send a JSON object")
			return
		}
		if req.ClientID != nil || req.ID != nil {
			AbortProblem(c, 422, "Client ID is immutable", "create a new Client or rotate its enrollment")
			return
		}
		updates := map[string]any{}
		if req.Comment != nil {
			updates["comment"] = strings.TrimSpace(*req.Comment)
		}
		if req.Enabled != nil {
			updates["enabled"] = *req.Enabled
			updates["stopped"] = !*req.Enabled
		}
		if req.LocationIPOverride != nil {
			value := strings.TrimSpace(*req.LocationIPOverride)
			if value != "" && publicLocationIP(value) == "" {
				AbortProblem(c, 400, "Invalid location IP", "enter a public IPv4 or IPv6 address, or leave blank for automatic location")
				return
			}
			updates["location_ip_override"] = value
		}
		if len(updates) == 0 {
			AbortProblem(c, 400, "No changes", "provide comment, enabled, or locationIpOverride")
			return
		}
		if err := db.Model(&models.Client{}).Where("client_id = ? AND user_id = ? AND tenant_id = ?", id, user.GetUserID(), user.GetTenantID()).Updates(updates).Error; err != nil {
			AbortProblem(c, 500, "Client update failed", "could not save the Client")
			return
		}
		db.Where("client_id = ?", id).First(&item)
		if req.Enabled != nil {
			go ReconcileClientTunnels(appInstance, id)
		}
		c.JSON(200, gin.H{"client": clientToResource(appInstance, db, &item)})
	}
}

func deleteClient(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		id := scopedID(user.GetUserName(), "c", c.Param("id"))
		db := appInstance.GetDBManager().GetDefaultDB()
		var item models.Client
		if db.Where("client_id = ? AND user_id = ? AND tenant_id = ? AND (origin_client_id IS NULL OR origin_client_id = ?)", id, user.GetUserID(), user.GetTenantID(), "").First(&item).Error != nil {
			AbortProblem(c, 404, "Client not found", "the requested Client does not exist")
			return
		}
		var deps []models.ProxyConfig
		db.Where("user_id = ? AND tenant_id = ? AND managed_by = ? AND (origin_client_id = ? OR client_id = ?)", user.GetUserID(), user.GetTenantID(), "tunnel", id, id).Find(&deps)
		if len(deps) > 0 {
			names := make([]gin.H, 0, len(deps))
			for _, d := range deps {
				names = append(names, gin.H{"id": d.PublicID, "name": d.Name})
			}
			AbortProblemWithDependencies(c, 409, "Client has dependent Tunnels", "delete or move the dependent Tunnels first", gin.H{"tunnels": names})
			return
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			// Enrollment tokens are one-use credentials. Remove any pending or
			// consumed record together with the physical Client so a deleted ID
			// can be recreated without leaving a stale unique-key reservation.
			if err := tx.Where("client_id = ? AND user_id = ? AND tenant_id = ?", id, user.GetUserID(), user.GetTenantID()).Delete(&models.AgentEnrollment{}).Error; err != nil {
				return err
			}
			return tx.Unscoped().Where("client_id = ? AND user_id = ? AND tenant_id = ?", id, user.GetUserID(), user.GetTenantID()).Delete(&models.Client{}).Error
		}); err != nil {
			AbortProblem(c, 500, "Client deletion failed", "could not delete the Client")
			return
		}
		if appInstance.GetClientsManager() != nil {
			appInstance.GetClientsManager().Remove(id)
		}
		c.Status(204)
	}
}

func rotateClientEnrollment(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		id := scopedID(user.GetUserName(), "c", c.Param("id"))
		db := appInstance.GetDBManager().GetDefaultDB()
		var item models.Client
		if db.Where("client_id = ? AND user_id = ? AND tenant_id = ? AND (origin_client_id IS NULL OR origin_client_id = ?)", id, user.GetUserID(), user.GetTenantID(), "").First(&item).Error != nil {
			AbortProblem(c, 404, "Client not found", "the requested Client does not exist")
			return
		}
		var req rotateEnrollmentRequest
		_ = c.ShouldBindJSON(&req)
		if isConfigured(item.EnrolledAt, item.ConfigContent) && !req.AcknowledgeDisruption {
			AbortProblem(c, 409, "Acknowledgement required", "rotating a configured Client credential requires acknowledgeDisruption=true")
			return
		}
		token, err := randomToken()
		if err != nil {
			AbortProblem(c, 500, "Enrollment rotation failed", "could not generate a secure token")
			return
		}
		expires := time.Now().UTC().Add(enrollmentLifetime)
		secret := utils.DeriveCredential(appInstance.GetConfig().App.GlobalSecret, "client-agent", token)
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("client_id = ?", id).Delete(&models.AgentEnrollment{}).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.Client{}).Where("client_id = ?", id).Updates(map[string]any{"connect_secret": utils.HashCredential(secret), "enrolled_at": nil, "last_seen_at": nil}).Error; err != nil {
				return err
			}
			return tx.Create(&models.AgentEnrollment{ID: uuid.NewString(), TokenHash: utils.HashCredential(token), ClientID: id, UserID: user.GetUserID(), TenantID: user.GetTenantID(), ExpiresAt: expires}).Error
		})
		if err != nil {
			AbortProblem(c, 500, "Enrollment rotation failed", "could not rotate the Client credential")
			return
		}
		if appInstance.GetClientsManager() != nil {
			appInstance.GetClientsManager().Remove(id)
		}
		db.Where("client_id = ?", id).First(&item)
		c.Header("Cache-Control", "no-store")
		c.JSON(200, gin.H{"client": clientToResource(appInstance, db, &item), "enrollment": makeEnrollment(appInstance, id, "client", 0, token, expires)})
	}
}

func createServer(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		var req serverCreateRequest
		if c.ShouldBindJSON(&req) != nil {
			AbortProblem(c, 400, "Invalid Server", "serverId, address, and bindPort are required")
			return
		}
		req.ServerID = strings.TrimSpace(req.ServerID)
		req.Address = strings.TrimSpace(req.Address)
		if !utils.IsClientIDPermited(req.ServerID) || !validServerHost(req.Address) {
			AbortProblem(c, 400, "Invalid Server", "use a valid Server ID and address")
			return
		}
		var portErr error
		req.BindPort, req.ServerAPIPort, portErr = normalizeServerPorts(req.BindPort, req.ServerAPIPort)
		if portErr != nil {
			AbortProblem(c, 400, "Invalid Server ports", portErr.Error())
			return
		}
		id := scopedID(user.GetUserName(), "s", req.ServerID)
		token, err := randomToken()
		if err != nil {
			AbortProblem(c, 500, "Server creation failed", "could not generate a secure enrollment token")
			return
		}
		expires := time.Now().UTC().Add(enrollmentLifetime)
		secret := utils.DeriveCredential(appInstance.GetConfig().App.GlobalSecret, "frps-server", token)
		item := models.Server{ServerEntity: &models.ServerEntity{ServerID: id, TenantID: user.GetTenantID(), UserID: user.GetUserID(), ServerIP: req.Address, BindPort: req.BindPort, ServerAPIPort: req.ServerAPIPort, Comment: strings.TrimSpace(req.Comment), ConnectSecret: utils.HashCredential(secret)}}
		if err := item.SetConfigContent(utils.NewBaseFRPServerUserAuthConfig(req.BindPort, nil)); err != nil {
			AbortProblem(c, 500, "Server creation failed", "could not create FRPS configuration")
			return
		}
		db := appInstance.GetDBManager().GetDefaultDB()
		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("server_id = ?", id).First(&models.Server{}).Error; err == nil {
				return gorm.ErrDuplicatedKey
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
			return tx.Create(&models.ServerEnrollment{ID: uuid.NewString(), TokenHash: utils.HashCredential(token), ServerID: id, UserID: user.GetUserID(), TenantID: user.GetTenantID(), ExpiresAt: expires}).Error
		})
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(strings.ToLower(err.Error()), "unique") {
				AbortProblem(c, 409, "Server already exists", "choose a different Server ID")
			} else {
				AbortProblem(c, 500, "Server creation failed", "could not save the Server")
			}
			return
		}
		c.Header("Cache-Control", "no-store")
		c.JSON(201, gin.H{"server": serverToResource(appInstance, db, &item), "enrollment": makeEnrollment(appInstance, id, "server", req.ServerAPIPort, token, expires)})
	}
}

func getServer(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		id := scopedID(user.GetUserName(), "s", c.Param("id"))
		var item models.Server
		if appInstance.GetDBManager().GetDefaultDB().Where("server_id = ? AND user_id = ? AND tenant_id = ?", id, user.GetUserID(), user.GetTenantID()).First(&item).Error != nil {
			AbortProblem(c, 404, "Server not found", "the requested Server does not exist")
			return
		}
		c.JSON(200, gin.H{"server": serverToResource(appInstance, appInstance.GetDBManager().GetDefaultDB(), &item)})
	}
}

func listServers(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		page, size := resourcePage(c)
		db := appInstance.GetDBManager().GetDefaultDB()
		q := db.Where("user_id = ? AND tenant_id = ?", user.GetUserID(), user.GetTenantID())
		if search := strings.TrimSpace(c.Query("search")); search != "" {
			q = q.Where(db.Where("server_id LIKE ?", "%"+search+"%").Or("server_ip LIKE ?", "%"+search+"%").Or("comment LIKE ?", "%"+search+"%"))
		}
		state := c.Query("configurationState")
		if state == "configured" {
			q = q.Where("enrolled_at IS NOT NULL")
		} else if state == "unconfigured" {
			q = q.Where("enrolled_at IS NULL")
		}
		status := c.Query("status")
		memoryStatus := status == "online" || status == "offline"
		if status == "pending" {
			q = q.Where("enrolled_at IS NULL")
		} else if status != "" && !memoryStatus {
			q = q.Where("1 = 0")
		}
		if memoryStatus {
			var all []models.Server
			if err := q.Order("created_at DESC").Find(&all).Error; err != nil {
				AbortProblem(c, 500, "Servers unavailable", "could not list Servers")
				return
			}
			items := make([]serverResource, 0, len(all))
			for i := range all {
				r := serverToResource(appInstance, db, &all[i])
				if r.Status == status {
					items = append(items, r)
				}
			}
			c.JSON(200, pageResult[serverResource]{Items: pageSlice(items, page, size), Total: int64(len(items)), Page: page, PageSize: size})
			return
		}
		var total int64
		if q.Model(&models.Server{}).Count(&total).Error != nil {
			AbortProblem(c, 500, "Servers unavailable", "could not count Servers")
			return
		}
		var rows []models.Server
		if q.Order("created_at DESC").Offset((page-1)*size).Limit(size).Find(&rows).Error != nil {
			AbortProblem(c, 500, "Servers unavailable", "could not list Servers")
			return
		}
		items := make([]serverResource, 0, len(rows))
		for i := range rows {
			r := serverToResource(appInstance, db, &rows[i])
			items = append(items, r)
		}
		c.JSON(200, pageResult[serverResource]{Items: items, Total: total, Page: page, PageSize: size})
	}
}

func patchServer(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		id := scopedID(user.GetUserName(), "s", c.Param("id"))
		db := appInstance.GetDBManager().GetDefaultDB()
		var item models.Server
		if db.Where("server_id = ? AND user_id = ? AND tenant_id = ?", id, user.GetUserID(), user.GetTenantID()).First(&item).Error != nil {
			AbortProblem(c, 404, "Server not found", "the requested Server does not exist")
			return
		}
		var req serverPatchRequest
		if c.ShouldBindJSON(&req) != nil {
			AbortProblem(c, 400, "Invalid Server update", "send a JSON object")
			return
		}
		if req.ServerID != nil || req.ID != nil {
			AbortProblem(c, 422, "Server ID is immutable", "create a new Server or rotate its enrollment")
			return
		}
		updates := map[string]any{}
		if req.Comment != nil {
			updates["comment"] = strings.TrimSpace(*req.Comment)
		}
		addressChanged := false
		bindPort := item.BindPort
		serverAPIPort := item.ServerAPIPort
		if bindPort == 0 {
			bindPort = defs.DefaultFRPSBindPort
		}
		if serverAPIPort == 0 {
			serverAPIPort = defs.DefaultServerAPIPort
		}
		if req.Address != nil {
			address := strings.TrimSpace(*req.Address)
			if !validServerHost(address) {
				AbortProblem(c, 400, "Invalid Server address", "use a public IP address or DNS hostname")
				return
			}
			updates["server_ip"] = address
			item.ServerIP = address
			addressChanged = true
		}
		if req.BindPort != nil {
			bindPort = *req.BindPort
			updates["bind_port"] = bindPort
			item.BindPort = bindPort
			addressChanged = true
		}
		if req.ServerAPIPort != nil {
			serverAPIPort = *req.ServerAPIPort
			updates["server_api_port"] = serverAPIPort
			item.ServerAPIPort = serverAPIPort
		}
		var portErr error
		bindPort, serverAPIPort, portErr = normalizeServerPorts(bindPort, serverAPIPort)
		if portErr != nil {
			AbortProblem(c, 400, "Invalid Server ports", portErr.Error())
			return
		}
		if req.BindPort != nil {
			updates["bind_port"] = bindPort
			item.BindPort = bindPort
		}
		if req.ServerAPIPort != nil {
			updates["server_api_port"] = serverAPIPort
			item.ServerAPIPort = serverAPIPort
		}
		if addressChanged {
			cfg, configErr := item.GetConfigContent()
			if configErr != nil || cfg == nil {
				cfg = utils.NewBaseFRPServerUserAuthConfig(item.BindPort, nil)
			}
			cfg.BindPort = item.BindPort
			raw, marshalErr := json.Marshal(cfg)
			if marshalErr != nil {
				AbortProblem(c, 500, "Server update failed", "could not encode the Server configuration")
				return
			}
			updates["config_content"] = raw
		}
		if len(updates) == 0 {
			AbortProblem(c, 400, "No changes", "provide address, bindPort, serverApiPort, or comment")
			return
		}
		if db.Model(&models.Server{}).Where("server_id = ? AND user_id = ? AND tenant_id = ?", id, user.GetUserID(), user.GetTenantID()).Updates(updates).Error != nil {
			AbortProblem(c, 500, "Server update failed", "could not save the Server")
			return
		}
		db.Where("server_id = ?", id).First(&item)
		if addressChanged {
			go func() {
				pushServerConfiguration(appInstance, id)
				ReconcileServerTunnels(appInstance, id)
			}()
		}
		c.JSON(200, gin.H{"server": serverToResource(appInstance, db, &item)})
	}
}

func deleteServer(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		id := scopedID(user.GetUserName(), "s", c.Param("id"))
		db := appInstance.GetDBManager().GetDefaultDB()
		var item models.Server
		if db.Where("server_id = ? AND user_id = ? AND tenant_id = ?", id, user.GetUserID(), user.GetTenantID()).First(&item).Error != nil {
			AbortProblem(c, 404, "Server not found", "the requested Server does not exist")
			return
		}
		var deps []models.ProxyConfig
		db.Where("user_id = ? AND tenant_id = ? AND managed_by = ? AND server_id = ?", user.GetUserID(), user.GetTenantID(), "tunnel", id).Find(&deps)
		if len(deps) > 0 {
			names := make([]gin.H, 0, len(deps))
			for _, d := range deps {
				names = append(names, gin.H{"id": d.PublicID, "name": d.Name})
			}
			AbortProblemWithDependencies(c, 409, "Server has dependent Tunnels", "delete or move the dependent Tunnels first", gin.H{"tunnels": names})
			return
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("server_id = ? AND user_id = ? AND tenant_id = ?", id, user.GetUserID(), user.GetTenantID()).Delete(&models.ServerEnrollment{}).Error; err != nil {
				return err
			}
			return tx.Unscoped().Where("server_id = ? AND user_id = ? AND tenant_id = ?", id, user.GetUserID(), user.GetTenantID()).Delete(&models.Server{}).Error
		}); err != nil {
			AbortProblem(c, 500, "Server deletion failed", "could not delete the Server")
			return
		}
		if appInstance.GetClientsManager() != nil {
			appInstance.GetClientsManager().Remove(id)
		}
		c.Status(204)
	}
}

func rotateServerEnrollment(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		id := scopedID(user.GetUserName(), "s", c.Param("id"))
		db := appInstance.GetDBManager().GetDefaultDB()
		var item models.Server
		if db.Where("server_id = ? AND user_id = ? AND tenant_id = ?", id, user.GetUserID(), user.GetTenantID()).First(&item).Error != nil {
			AbortProblem(c, 404, "Server not found", "the requested Server does not exist")
			return
		}
		var req rotateEnrollmentRequest
		_ = c.ShouldBindJSON(&req)
		if isConfigured(item.EnrolledAt, item.ConfigContent) && !req.AcknowledgeDisruption {
			AbortProblem(c, 409, "Acknowledgement required", "rotating a configured Server credential requires acknowledgeDisruption=true")
			return
		}
		token, err := randomToken()
		if err != nil {
			AbortProblem(c, 500, "Enrollment rotation failed", "could not generate a secure token")
			return
		}
		expires := time.Now().UTC().Add(enrollmentLifetime)
		secret := utils.DeriveCredential(appInstance.GetConfig().App.GlobalSecret, "frps-server", token)
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("server_id = ?", id).Delete(&models.ServerEnrollment{}).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.Server{}).Where("server_id = ?", id).Updates(map[string]any{"connect_secret": utils.HashCredential(secret), "enrolled_at": nil, "last_seen_at": nil}).Error; err != nil {
				return err
			}
			return tx.Create(&models.ServerEnrollment{ID: uuid.NewString(), TokenHash: utils.HashCredential(token), ServerID: id, UserID: user.GetUserID(), TenantID: user.GetTenantID(), ExpiresAt: expires}).Error
		}); err != nil {
			AbortProblem(c, 500, "Enrollment rotation failed", "could not rotate the Server credential")
			return
		}
		if appInstance.GetClientsManager() != nil {
			appInstance.GetClientsManager().Remove(id)
		}
		db.Where("server_id = ?", id).First(&item)
		c.Header("Cache-Control", "no-store")
		c.JSON(200, gin.H{"server": serverToResource(appInstance, db, &item), "enrollment": makeEnrollment(appInstance, id, "server", item.ServerAPIPort, token, expires)})
	}
}

func normalizeTunnelRequest(user models.UserInfo, req *tunnelRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	req.ClientID = scopedID(user.GetUserName(), "c", req.ClientID)
	req.ServerID = scopedID(user.GetUserName(), "s", req.ServerID)
	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	req.LocalHost = strings.TrimSpace(req.LocalHost)
	if req.LocalHost == "" {
		req.LocalHost = "127.0.0.1"
	}
	if !utils.IsClientIDPermited(req.Name) {
		return fmt.Errorf("name must use letters, numbers, underscores, or hyphens")
	}
	if req.Type != "tcp" && req.Type != "udp" {
		return fmt.Errorf("type must be tcp or udp")
	}
	if !validServerHost(req.LocalHost) || req.LocalPort < 1 || req.LocalPort > 65535 || req.RemotePort < 1024 || req.RemotePort > 65535 {
		return fmt.Errorf("invalid local or remote endpoint")
	}
	return nil
}

func tunnelContent(req tunnelRequest) ([]byte, error) {
	base := v1.ProxyBaseConfig{Name: req.Name, Type: req.Type, ProxyBackend: v1.ProxyBackend{LocalIP: req.LocalHost, LocalPort: req.LocalPort}}
	var p v1.ProxyConfigurer
	if req.Type == "tcp" {
		p = &v1.TCPProxyConfig{ProxyBaseConfig: base, RemotePort: req.RemotePort}
	} else {
		p = &v1.UDPProxyConfig{ProxyBaseConfig: base, RemotePort: req.RemotePort}
	}
	return json.Marshal(p)
}

// ensureTunnelRemotePortAvailable prevents a desired Tunnel from colliding
// with another active listener on the same Server. TCP and UDP may share a
// numeric port, but two listeners of the same protocol cannot.
func ensureTunnelRemotePortAvailable(db *gorm.DB, user models.UserInfo, req tunnelRequest, excludePublicID string) error {
	if req.Enabled != nil && *req.Enabled {
		var server models.Server
		if err := db.Where("server_id = ? AND user_id = ? AND tenant_id = ?", req.ServerID, user.GetUserID(), user.GetTenantID()).First(&server).Error; err != nil {
			return fmt.Errorf("check remote port availability: %w", err)
		}
		bindPort := server.BindPort
		if bindPort == 0 {
			bindPort = defs.DefaultFRPSBindPort
		}
		serverAPIPort := server.ServerAPIPort
		if serverAPIPort == 0 {
			serverAPIPort = defs.DefaultServerAPIPort
		}
		if req.RemotePort == bindPort {
			return fmt.Errorf("remote port %d is reserved by the Server Bind port", req.RemotePort)
		}
		if req.RemotePort == serverAPIPort {
			return fmt.Errorf("remote port %d is reserved by SERVER_API_PORT", req.RemotePort)
		}
	}
	var rows []models.ProxyConfig
	q := db.Where("user_id = ? AND tenant_id = ? AND managed_by = ? AND server_id = ? AND stopped = ?", user.GetUserID(), user.GetTenantID(), "tunnel", req.ServerID, false)
	if excludePublicID != "" {
		q = q.Where("public_id <> ?", excludePublicID)
	}
	if err := q.Find(&rows).Error; err != nil {
		return fmt.Errorf("check remote port availability: %w", err)
	}
	for i := range rows {
		other := tunnelFromModel(&rows[i])
		if other.Type == req.Type && other.RemotePort == req.RemotePort {
			return fmt.Errorf("remote port %d is already used by Tunnel %q on this Server", req.RemotePort, rows[i].Name)
		}
	}
	return nil
}

func createTunnelResource(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		var req tunnelRequest
		if c.ShouldBindJSON(&req) != nil || normalizeTunnelRequest(user, &req) != nil {
			AbortProblem(c, 400, "Invalid Tunnel", "name, clientId, serverId, type, localPort, and remotePort are required")
			return
		}
		if req.Enabled == nil {
			req.Enabled = boolPtr(true)
		}
		db := appInstance.GetDBManager().GetDefaultDB()
		var client models.Client
		var server models.Server
		if db.Where("client_id = ? AND user_id = ? AND tenant_id = ? AND (origin_client_id IS NULL OR origin_client_id = ?)", req.ClientID, user.GetUserID(), user.GetTenantID(), "").First(&client).Error != nil {
			AbortProblem(c, 404, "Client not found", "select an existing Client")
			return
		}
		if db.Where("server_id = ? AND user_id = ? AND tenant_id = ?", req.ServerID, user.GetUserID(), user.GetTenantID()).First(&server).Error != nil {
			AbortProblem(c, 404, "Server not found", "select an existing Server")
			return
		}
		lock, _ := tunnelServerLocks.LoadOrStore(req.ServerID, &sync.Mutex{})
		mu := lock.(*sync.Mutex)
		mu.Lock()
		defer mu.Unlock()
		var count int64
		db.Model(&models.ProxyConfig{}).Where("user_id = ? AND tenant_id = ? AND managed_by = ? AND origin_client_id = ? AND name = ?", user.GetUserID(), user.GetTenantID(), "tunnel", req.ClientID, req.Name).Count(&count)
		if count > 0 {
			AbortProblem(c, 409, "Tunnel already exists", "choose a unique Tunnel name for this Client")
			return
		}
		if err := ensureTunnelRemotePortAvailable(db, user, req, ""); err != nil {
			if strings.HasPrefix(err.Error(), "check remote port availability:") {
				AbortProblem(c, 500, "Tunnel validation failed", "could not check the Server remote port")
			} else {
				AbortProblem(c, 409, "Remote port already in use", err.Error())
			}
			return
		}
		content, err := tunnelContent(req)
		if err != nil {
			AbortProblem(c, 400, "Invalid Tunnel", "could not encode the Tunnel configuration")
			return
		}
		item := &models.ProxyConfig{ProxyConfigEntity: &models.ProxyConfigEntity{PublicID: uuid.NewString(), ServerID: req.ServerID, ClientID: req.ClientID, OriginClientID: req.ClientID, Name: req.Name, Type: req.Type, UserID: user.GetUserID(), TenantID: user.GetTenantID(), Content: content, Stopped: !*req.Enabled, ManagedBy: "tunnel", DesiredRevision: 1}}
		if err := db.Create(item).Error; err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") {
				AbortProblem(c, 409, "Tunnel already exists", "choose a unique Tunnel name for this Client")
			} else {
				AbortProblem(c, 500, "Tunnel creation failed", "could not save the Tunnel")
			}
			return
		}
		go reconcileTunnel(appInstance, item.ID)
		c.JSON(201, gin.H{"tunnel": tunnelToResource(appInstance, db, item)})
	}
}

func listTunnelsResource(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		page, size := resourcePage(c)
		db := appInstance.GetDBManager().GetDefaultDB()
		q := db.Where("user_id = ? AND tenant_id = ? AND managed_by = ?", user.GetUserID(), user.GetTenantID(), "tunnel")
		if search := strings.TrimSpace(c.Query("search")); search != "" {
			q = q.Where(db.Where("name LIKE ?", "%"+search+"%").Or("origin_client_id LIKE ?", "%"+search+"%").Or("server_id LIKE ?", "%"+search+"%"))
		}
		status := c.Query("status")
		if status != "" {
			var all []models.ProxyConfig
			if err := q.Order("updated_at DESC").Find(&all).Error; err != nil {
				AbortProblem(c, 500, "Tunnels unavailable", "could not list Tunnels")
				return
			}
			items := make([]tunnelResource, 0, len(all))
			for i := range all {
				r := tunnelToResource(appInstance, db, &all[i])
				if r.Status == status {
					items = append(items, r)
				}
			}
			c.JSON(200, pageResult[tunnelResource]{Items: pageSlice(items, page, size), Total: int64(len(items)), Page: page, PageSize: size})
			return
		}
		var total int64
		if q.Model(&models.ProxyConfig{}).Count(&total).Error != nil {
			AbortProblem(c, 500, "Tunnels unavailable", "could not count Tunnels")
			return
		}
		var rows []models.ProxyConfig
		if q.Order("updated_at DESC").Offset((page-1)*size).Limit(size).Find(&rows).Error != nil {
			AbortProblem(c, 500, "Tunnels unavailable", "could not list Tunnels")
			return
		}
		items := make([]tunnelResource, 0, len(rows))
		for i := range rows {
			r := tunnelToResource(appInstance, db, &rows[i])
			items = append(items, r)
		}
		c.JSON(200, pageResult[tunnelResource]{Items: items, Total: total, Page: page, PageSize: size})
	}
}

func getTunnel(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		db := appInstance.GetDBManager().GetDefaultDB()
		var item models.ProxyConfig
		if db.Where("public_id = ? AND user_id = ? AND tenant_id = ? AND managed_by = ?", c.Param("id"), user.GetUserID(), user.GetTenantID(), "tunnel").First(&item).Error != nil {
			AbortProblem(c, 404, "Tunnel not found", "the requested Tunnel does not exist")
			return
		}
		c.JSON(200, gin.H{"tunnel": tunnelToResource(appInstance, db, &item)})
	}
}

func patchTunnel(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		db := appInstance.GetDBManager().GetDefaultDB()
		var item models.ProxyConfig
		if db.Where("public_id = ? AND user_id = ? AND tenant_id = ? AND managed_by = ?", c.Param("id"), user.GetUserID(), user.GetTenantID(), "tunnel").First(&item).Error != nil {
			AbortProblem(c, 404, "Tunnel not found", "the requested Tunnel does not exist")
			return
		}
		var req tunnelPatchRequest
		if c.ShouldBindJSON(&req) != nil {
			AbortProblem(c, 400, "Invalid Tunnel update", "send a JSON object")
			return
		}
		previousClientID, previousServerID := item.OriginClientID, item.ServerID
		old := tunnelFromModel(&item)
		if req.Name != nil {
			old.Name = *req.Name
		}
		if req.ClientID != nil {
			old.ClientID = *req.ClientID
		}
		if req.ServerID != nil {
			old.ServerID = *req.ServerID
		}
		if req.Type != nil {
			old.Type = *req.Type
		}
		if req.LocalHost != nil {
			old.LocalHost = *req.LocalHost
		}
		if req.LocalPort != nil {
			old.LocalPort = *req.LocalPort
		}
		if req.RemotePort != nil {
			old.RemotePort = *req.RemotePort
		}
		if req.Enabled != nil {
			old.Enabled = req.Enabled
		}
		if old.Enabled == nil {
			old.Enabled = boolPtr(!item.Stopped)
		}
		if err := normalizeTunnelRequest(user, &old); err != nil {
			AbortProblem(c, 400, "Invalid Tunnel update", err.Error())
			return
		}
		lock, _ := tunnelServerLocks.LoadOrStore(old.ServerID, &sync.Mutex{})
		mu := lock.(*sync.Mutex)
		mu.Lock()
		defer mu.Unlock()
		if old.ClientID != item.OriginClientID || old.Name != item.Name {
			var count int64
			db.Model(&models.ProxyConfig{}).Where("user_id = ? AND tenant_id = ? AND managed_by = ? AND origin_client_id = ? AND name = ? AND public_id <> ?", user.GetUserID(), user.GetTenantID(), "tunnel", old.ClientID, old.Name, item.PublicID).Count(&count)
			if count > 0 {
				AbortProblem(c, 409, "Tunnel already exists", "choose a unique Tunnel name for this Client")
				return
			}
		}
		if err := ensureTunnelRemotePortAvailable(db, user, old, item.PublicID); err != nil {
			if strings.HasPrefix(err.Error(), "check remote port availability:") {
				AbortProblem(c, 500, "Tunnel validation failed", "could not check the Server remote port")
			} else {
				AbortProblem(c, 409, "Remote port already in use", err.Error())
			}
			return
		}
		content, err := tunnelContent(old)
		if err != nil {
			AbortProblem(c, 400, "Invalid Tunnel update", err.Error())
			return
		}
		nextRevision := item.DesiredRevision + 1
		updates := map[string]any{"name": old.Name, "type": old.Type, "server_id": old.ServerID, "client_id": old.ClientID, "origin_client_id": old.ClientID, "content": content, "stopped": !*old.Enabled, "desired_revision": nextRevision, "last_error": ""}
		if err := db.Model(&models.ProxyConfig{}).Where("public_id = ?", item.PublicID).Updates(updates).Error; err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") {
				AbortProblem(c, 409, "Tunnel already exists", "choose a unique Tunnel name for this Client")
			} else {
				AbortProblem(c, 500, "Tunnel update failed", "could not save the Tunnel")
			}
			return
		}
		db.Where("public_id = ?", item.PublicID).First(&item)
		go applyTunnelPair(appInstance, previousClientID, previousServerID)
		if previousClientID != item.OriginClientID || previousServerID != item.ServerID {
			go applyTunnelPair(appInstance, item.OriginClientID, item.ServerID)
		}
		c.JSON(200, gin.H{"tunnel": tunnelToResource(appInstance, db, &item)})
	}
}

func deleteTunnelResource(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		db := appInstance.GetDBManager().GetDefaultDB()
		var item models.ProxyConfig
		if db.Where("public_id = ? AND user_id = ? AND tenant_id = ? AND managed_by = ?", c.Param("id"), user.GetUserID(), user.GetTenantID(), "tunnel").First(&item).Error != nil {
			AbortProblem(c, 404, "Tunnel not found", "the requested Tunnel does not exist")
			return
		}
		oldClientID, oldServerID := item.OriginClientID, item.ServerID
		if err := db.Delete(&item).Error; err != nil {
			AbortProblem(c, 500, "Tunnel deletion failed", "could not delete the Tunnel")
			return
		}
		go applyTunnelPair(appInstance, oldClientID, oldServerID)
		c.Status(204)
	}
}

func overview(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		db := appInstance.GetDBManager().GetDefaultDB()
		var clients, servers, tunnels int64
		db.Model(&models.Client{}).Where("user_id = ? AND tenant_id = ? AND (origin_client_id IS NULL OR origin_client_id = ?)", user.GetUserID(), user.GetTenantID(), "").Count(&clients)
		db.Model(&models.Server{}).Where("user_id = ? AND tenant_id = ?", user.GetUserID(), user.GetTenantID()).Count(&servers)
		db.Model(&models.ProxyConfig{}).Where("user_id = ? AND tenant_id = ? AND managed_by = ?", user.GetUserID(), user.GetTenantID(), "tunnel").Count(&tunnels)
		c.JSON(200, gin.H{"clients": clients, "servers": servers, "tunnels": tunnels})
	}
}

func publicLocationIP(value string) string {
	ip := net.ParseIP(strings.Trim(strings.TrimSpace(value), "[]"))
	if ip == nil || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast() {
		return ""
	}
	return ip.String()
}

func clientLocation(item *models.ClientEntity) (string, string) {
	if ip := publicLocationIP(item.LocationIPOverride); ip != "" {
		return ip, "manual"
	}
	if ip := publicLocationIP(item.ReportedIP); ip != "" && item.ReportedAt != nil && time.Since(*item.ReportedAt) < 24*time.Hour {
		return ip, "agent_probe"
	}
	if ip := publicLocationIP(item.LastSeenIP); ip != "" {
		return ip, "observed"
	}
	return "", ""
}

func topology(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			AbortProblem(c, 401, "Unauthorized", "a valid user session is required")
			return
		}
		db := appInstance.GetDBManager().GetDefaultDB()
		var clients []models.Client
		if err := db.Where("user_id = ? AND tenant_id = ? AND (origin_client_id IS NULL OR origin_client_id = ?)", user.GetUserID(), user.GetTenantID(), "").Order("created_at DESC").Find(&clients).Error; err != nil {
			AbortProblem(c, 500, "Topology unavailable", "could not list Clients")
			return
		}
		var servers []models.Server
		if err := db.Where("user_id = ? AND tenant_id = ?", user.GetUserID(), user.GetTenantID()).Order("created_at DESC").Find(&servers).Error; err != nil {
			AbortProblem(c, 500, "Topology unavailable", "could not list Servers")
			return
		}
		var tunnels []models.ProxyConfig
		if err := db.Where("user_id = ? AND tenant_id = ? AND managed_by = ?", user.GetUserID(), user.GetTenantID(), "tunnel").Order("updated_at DESC").Find(&tunnels).Error; err != nil {
			AbortProblem(c, 500, "Topology unavailable", "could not list Tunnels")
			return
		}

		nodes := make([]topologyNode, 0, len(clients)+len(servers))
		clientIDs := make(map[string]struct{}, len(clients))
		serverIDs := make(map[string]struct{}, len(servers))
		for i := range clients {
			item := &clients[i]
			resource := clientToResource(appInstance, db, item)
			clientIDs[item.ClientID] = struct{}{}
			locationIP, source := clientLocation(item.ClientEntity)
			nodes = append(nodes, topologyNode{
				ID: item.ClientID, Kind: "client", Label: item.ClientID, Comment: item.Comment,
				Status: resource.Status, ConfigurationState: resource.ConfigurationState,
				Enabled: resource.Enabled, LocationIP: locationIP, LocationSource: source,
				ObservedIP: publicLocationIP(item.LastSeenIP), ReportedIP: publicLocationIP(item.ReportedIP), ReportedAt: item.ReportedAt,
				LastSeenAt: item.LastSeenAt, TunnelCount: resource.TunnelCount,
			})
		}
		for i := range servers {
			item := &servers[i]
			resource := serverToResource(appInstance, db, item)
			serverIDs[item.ServerID] = struct{}{}
			locationIP := publicLocationIP(item.ServerIP)
			locationSource := "configured"
			if locationIP == "" {
				locationIP = publicLocationIP(item.LastSeenIP)
				locationSource = "observed"
			}
			nodes = append(nodes, topologyNode{
				ID: item.ServerID, Kind: "server", Label: item.ServerID, Comment: item.Comment,
				Address: item.ServerIP, Status: resource.Status, ConfigurationState: resource.ConfigurationState,
				Enabled: true, LocationIP: locationIP, LocationSource: locationSource, ObservedIP: publicLocationIP(item.LastSeenIP), LastSeenAt: item.LastSeenAt, TunnelCount: resource.TunnelCount,
			})
		}

		links := make([]topologyLink, 0, len(tunnels))
		for i := range tunnels {
			item := &tunnels[i]
			if _, ok := clientIDs[item.OriginClientID]; !ok {
				continue
			}
			if _, ok := serverIDs[item.ServerID]; !ok {
				continue
			}
			resource := tunnelToResource(appInstance, db, item)
			links = append(links, topologyLink{
				ID: resource.ID, Name: resource.Name, SourceClientID: resource.ClientID,
				TargetServerID: resource.ServerID, Type: resource.Type, RemotePort: resource.RemotePort,
				Enabled: resource.Enabled, Status: resource.Status, LastError: resource.LastError,
			})
		}
		located := 0
		for _, node := range nodes {
			if node.LocationIP != "" {
				located++
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"nodes": nodes, "links": links, "locatedCount": located,
			"totalCount": len(nodes), "generatedAt": time.Now().UTC(),
		})
	}
}

func tunnelFromModel(item *models.ProxyConfig) tunnelRequest {
	r := tunnelRequest{Name: item.Name, ClientID: item.OriginClientID, ServerID: item.ServerID, Type: item.Type, Enabled: boolPtr(!item.Stopped)}
	var cfg v1.TypedProxyConfig
	if err := cfg.UnmarshalJSON(item.Content); err == nil && cfg.GetBaseConfig() != nil {
		r.LocalHost = cfg.GetBaseConfig().LocalIP
		r.LocalPort = cfg.GetBaseConfig().LocalPort
		switch p := cfg.ProxyConfigurer.(type) {
		case *v1.TCPProxyConfig:
			r.RemotePort = p.RemotePort
		case *v1.UDPProxyConfig:
			r.RemotePort = p.RemotePort
		}
	}
	return r
}
func boolPtr(value bool) *bool { return &value }
func tunnelToResource(appInstance app.Application, db *gorm.DB, item *models.ProxyConfig) tunnelResource {
	r := tunnelFromModel(item)
	serverAddress := ""
	if db != nil && item.ServerID != "" {
		var server models.Server
		if err := db.Where("server_id = ? AND user_id = ? AND tenant_id = ?", item.ServerID, item.UserID, item.TenantID).First(&server).Error; err == nil && server.ServerEntity != nil {
			serverAddress = server.ServerIP
		}
	}
	return tunnelResource{ID: item.PublicID, Name: item.Name, ClientID: item.OriginClientID, ServerID: item.ServerID, ServerAddress: serverAddress, Type: item.Type, LocalHost: r.LocalHost, LocalPort: r.LocalPort, RemotePort: r.RemotePort, Enabled: !item.Stopped, Status: tunnelStatus(appInstance, item), LastError: item.LastError, UpdatedAt: item.UpdatedAt}
}

func makeEnrollment(appInstance app.Application, id, kind string, serverAPIPort int, token string, expires time.Time) enrollmentPayload {
	cfg := appInstance.GetConfig()
	apiURL := conf.GetAPIURL(cfg)
	rpcURL := cfg.Client.RPCUrl
	if rpcURL == "" {
		rpcURL = apiURL
	}
	p := enrollmentPayload{Token: token, ExpiresAt: expires, APIURL: apiURL, RPCURL: rpcURL}
	if kind == "client" {
		p.ClientID = id
		base := strings.TrimRight(cfg.App.AgentInstallURL, "/")
		if base == "" {
			base = "https://raw.githubusercontent.com/Onicc/frp-panel/main"
		}
		args := fmt.Sprintf("--client-id %s --enrollment-token %s --api-url %s --rpc-url %s", shellQuote(id), shellQuote(token), shellQuote(apiURL), shellQuote(rpcURL))
		p.InstallCommands = map[string]string{
			"linux":   fmt.Sprintf("curl --fail --silent --show-error --location --proto '=https' --tlsv1.2 %s | sudo bash -s -- %s", shellQuote(base+"/install.sh"), args),
			"darwin":  fmt.Sprintf("curl --fail --silent --show-error --location --proto '=https' --tlsv1.2 %s | sudo bash -s -- %s", shellQuote(base+"/install.sh"), args),
			"windows": powershellInstallCommand(base, id, token, apiURL, rpcURL),
		}
		// Keep the Linux command as a compatibility field for existing clients.
		p.InstallCommand = p.InstallCommands["linux"]
	} else {
		p.ServerID = id
		if serverAPIPort == 0 {
			serverAPIPort = defs.DefaultServerAPIPort
		}
		p.ComposeYAML = fmt.Sprintf(`services:
  frps:
    image: ${FRP_PANEL_IMAGE:-onicc/frp-panel:edge}
    restart: unless-stopped
    network_mode: host
    command: ["server", "--config", "/data/server.yaml"]
    environment:
      PUBLIC_URL: %q
      SERVER_ENROLLMENT_TOKEN: %q
      SERVER_API_PORT: %q
    volumes:
      - frp-panel-server-data:/data

volumes:
  frp-panel-server-data:
`, apiURL, token, strconv.Itoa(serverAPIPort))
	}
	return p
}

// shellQuote returns a single-quoted POSIX shell argument. Enrollment values
// are generated by Master, but quoting here keeps the copied command safe even
// when a deployment uses a custom installation base URL.
func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func powershellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func powershellInstallCommand(base, id, token, apiURL, rpcURL string) string {
	args := []string{
		"'--client-id'", powershellQuote(id),
		"'--enrollment-token'", powershellQuote(token),
		"'--api-url'", powershellQuote(apiURL),
		"'--rpc-url'", powershellQuote(rpcURL),
	}
	return fmt.Sprintf("powershell -NoProfile -ExecutionPolicy Bypass -Command \"$p = Join-Path $env:TEMP ('frp-panel-install-' + [guid]::NewGuid().ToString('N') + '.ps1'); try { [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -UseBasicParsing -Uri %s -OutFile $p; & $p -AgentArguments @(%s); if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE } } finally { Remove-Item -LiteralPath $p -Force -ErrorAction SilentlyContinue }\"", powershellQuote(strings.TrimRight(base, "/")+"/install.ps1"), strings.Join(args, ", "))
}

// reconcileTunnel is deliberately best-effort. Desired state is durable and
// can be applied when either endpoint reconnects; failures are recorded for
// operators instead of turning a successful CRUD request into a timeout.
func reconcileTunnel(appInstance app.Application, rowID uint) {
	db := appInstance.GetDBManager().GetDefaultDB()
	var item models.ProxyConfig
	if db.First(&item, rowID).Error != nil {
		return
	}
	if item.ManagedBy != "tunnel" {
		return
	}
	applyTunnelPair(appInstance, item.OriginClientID, item.ServerID)
}

// ReconcileClientTunnels re-applies durable Tunnel state after an Agent
// reconnects or its desired lifecycle state changes.
func ReconcileClientTunnels(appInstance app.Application, clientID string) {
	var rows []models.ProxyConfig
	db := appInstance.GetDBManager().GetDefaultDB()
	if db.Where("managed_by = ? AND origin_client_id = ?", "tunnel", clientID).Find(&rows).Error != nil {
		return
	}
	seen := map[string]bool{}
	for _, row := range rows {
		if !seen[row.ServerID] {
			seen[row.ServerID] = true
			applyTunnelPair(appInstance, clientID, row.ServerID)
		}
	}
}

// ResetAndReconcileClient removes runtime connections that may no longer have
// database rows, then rebuilds every currently desired pair on reconnect.
func ResetAndReconcileClient(appInstance app.Application, clientID string) {
	if !connectorOnline(appInstance, clientID) {
		return
	}
	ctx := app.NewContext(context.Background(), appInstance)
	if _, err := rpc.CallClient(ctx, clientID, pb.Event_EVENT_REMOVE_FRPC, &pb.RemoveFRPCRequest{ClientId: &clientID}); err != nil {
		appInstance.Logger(ctx).WithError(err).Warnf("could not reset Client runtime: %s", clientID)
		return
	}
	ReconcileClientTunnels(appInstance, clientID)
}

// ReconcileServerTunnels reapplies all Client connections for a Server after
// FRPS registration or a bind/address change.
func ReconcileServerTunnels(appInstance app.Application, serverID string) {
	var rows []models.ProxyConfig
	db := appInstance.GetDBManager().GetDefaultDB()
	if db.Where("managed_by = ? AND server_id = ?", "tunnel", serverID).Find(&rows).Error != nil {
		return
	}
	seen := map[string]bool{}
	for _, row := range rows {
		if !seen[row.OriginClientID] {
			seen[row.OriginClientID] = true
			applyTunnelPair(appInstance, row.OriginClientID, serverID)
		}
	}
}

func pushServerConfiguration(appInstance app.Application, serverID string) {
	if !connectorOnline(appInstance, serverID) {
		return
	}
	db := appInstance.GetDBManager().GetDefaultDB()
	var server models.Server
	if db.Where("server_id = ?", serverID).First(&server).Error != nil || len(server.ConfigContent) == 0 {
		return
	}
	request := &pb.UpdateFRPSRequest{
		ServerId: &serverID,
		Config:   server.ConfigContent,
		Comment:  &server.Comment,
		ServerIp: &server.ServerIP,
	}
	ctx := app.NewContext(context.Background(), appInstance)
	if _, err := rpc.CallClient(ctx, serverID, pb.Event_EVENT_UPDATE_FRPS, request); err != nil {
		appInstance.Logger(ctx).WithError(err).Warnf("could not push Server configuration, server id: [%s]", serverID)
	}
}

func applyTunnelPair(appInstance app.Application, clientID, serverID string) {
	lock, _ := tunnelPairLocks.LoadOrStore(clientID+"\x00"+serverID, &sync.Mutex{})
	mu := lock.(*sync.Mutex)
	mu.Lock()
	defer mu.Unlock()

	db := appInstance.GetDBManager().GetDefaultDB()
	var client models.Client
	var server models.Server
	if db.Where("client_id = ?", clientID).First(&client).Error != nil || db.Where("server_id = ?", serverID).First(&server).Error != nil {
		return
	}
	var rows []models.ProxyConfig
	if err := db.Where("managed_by = ? AND origin_client_id = ? AND server_id = ? AND stopped = ?", "tunnel", clientID, serverID, false).Order("id ASC").Find(&rows).Error; err != nil {
		return
	}
	if !connectorOnline(appInstance, clientID) {
		return
	}
	if len(rows) == 0 || !client.Enabled || client.Stopped || !isConfigured(client.EnrolledAt, client.ConfigContent) || !isConfigured(server.EnrolledAt, server.ConfigContent) {
		// An empty configuration addresses exactly one Client→Server connection.
		// REMOVE_FRPC addresses the whole physical Client and must not be used here.
		raw := []byte(`{"proxies":[]}`)
		ctx := app.NewContext(context.Background(), appInstance)
		_, _ = rpc.CallClient(ctx, clientID, pb.Event_EVENT_UPDATE_FRPC, &pb.UpdateFRPCRequest{ClientId: &clientID, ServerId: &serverID, Config: raw})
		return
	}
	var err error
	clientConfig := &v1.ClientConfig{}
	if len(client.ConfigContent) > 0 {
		clientConfig, err = client.GetConfigContent()
		if err != nil || clientConfig == nil {
			return
		}
	}
	serverConfig, err := server.GetConfigContent()
	if err != nil || serverConfig == nil {
		return
	}
	user := models.User{}
	if db.Where("user_id = ?", client.UserID).First(&user).Error != nil {
		return
	}
	bindPort := server.BindPort
	if bindPort == 0 {
		bindPort = serverConfig.BindPort
	}
	clientConfig.ClientCommonConfig = *utils.NewBaseFRPClientUserAuthConfig(server.ServerIP, bindPort, utils.FRPClientUser(user.UserName, clientID), user.Token)
	clientConfig.Proxies = nil
	for _, row := range rows {
		var typed v1.TypedProxyConfig
		if err := typed.UnmarshalJSON(row.Content); err != nil {
			setTunnelError(db, row.ID, err)
			continue
		}
		clientConfig.Proxies = append(clientConfig.Proxies, typed)
	}
	raw, err := json.Marshal(clientConfig)
	if err != nil {
		for _, row := range rows {
			setTunnelError(db, row.ID, err)
		}
		return
	}
	ctx := app.NewContext(context.Background(), appInstance)
	if _, err := rpc.CallClient(ctx, clientID, pb.Event_EVENT_UPDATE_FRPC, &pb.UpdateFRPCRequest{ClientId: &clientID, ServerId: &serverID, Config: raw}); err != nil {
		for _, row := range rows {
			setTunnelError(db, row.ID, err)
		}
		return
	}
	now := time.Now().UTC()
	for _, row := range rows {
		db.Model(&models.ProxyConfig{}).Where("id = ?", row.ID).Updates(map[string]any{"last_applied_at": now, "last_error": ""})
	}
}

func setTunnelError(db *gorm.DB, rowID uint, err error) {
	if err == nil {
		return
	}
	db.Model(&models.ProxyConfig{}).Where("id = ?", rowID).Update("last_error", err.Error())
}

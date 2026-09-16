package v2

import (
	"net/http"
	"strings"

	masterproxy "github.com/Onicc/frp-panel/biz/master/proxy"
	"github.com/Onicc/frp-panel/common"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/dao"
	"github.com/Onicc/frp-panel/utils"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/gin-gonic/gin"
)

type createTunnelRequest struct {
	Name       string `json:"name" binding:"required"`
	NodeID     string `json:"nodeId" binding:"required"`
	ServerID   string `json:"serverId" binding:"required"`
	Type       string `json:"type" binding:"required"`
	LocalHost  string `json:"localHost"`
	LocalPort  int    `json:"localPort" binding:"required"`
	RemotePort int    `json:"remotePort" binding:"required"`
}

type deleteTunnelRequest struct {
	Name     string `json:"name" binding:"required"`
	NodeID   string `json:"nodeId" binding:"required"`
	ServerID string `json:"serverId" binding:"required"`
}

type tunnelResponse struct {
	Name       string `json:"name"`
	NodeID     string `json:"nodeId"`
	ClientID   string `json:"clientId"`
	ServerID   string `json:"serverId"`
	Type       string `json:"type"`
	LocalHost  string `json:"localHost"`
	LocalPort  int    `json:"localPort"`
	RemotePort int    `json:"remotePort"`
	Stopped    bool   `json:"stopped"`
}

func createTunnel(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request createTunnelRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			AbortProblem(c, http.StatusBadRequest, "Invalid tunnel", "name, nodeId, serverId, type, localPort, and remotePort are required")
			return
		}
		request.Name = strings.TrimSpace(request.Name)
		request.LocalHost = strings.TrimSpace(request.LocalHost)
		if request.LocalHost == "" {
			request.LocalHost = "127.0.0.1"
		}
		if !utils.IsClientIDPermited(request.Name) {
			AbortProblem(c, http.StatusBadRequest, "Invalid tunnel name", "use letters, numbers, underscores, or hyphens")
			return
		}
		if request.Type != string(v1.ProxyTypeTCP) && request.Type != string(v1.ProxyTypeUDP) {
			AbortProblem(c, http.StatusBadRequest, "Unsupported tunnel type", "type must be tcp or udp")
			return
		}
		if !validServerHost(request.LocalHost) || request.LocalPort < 1 || request.LocalPort > 65535 || request.RemotePort < 1024 || request.RemotePort > 65535 {
			AbortProblem(c, http.StatusBadRequest, "Invalid tunnel endpoint", "use a valid local host, local port 1-65535, and remote port 1024-65535")
			return
		}

		userInfo := common.GetUserInfo(c)
		if userInfo == nil || !userInfo.Valid() {
			AbortProblem(c, http.StatusUnauthorized, "Unauthorized", "a valid user session is required")
			return
		}
		ctx := app.NewContext(c, appInstance)
		origin, err := dao.NewQuery(ctx).GetClientByClientID(userInfo, request.NodeID)
		if err != nil || origin.OriginClientID != "" {
			AbortProblem(c, http.StatusNotFound, "Node not found", "the requested managed node does not exist")
			return
		}
		child, err := dao.NewQuery(ctx).GetClientByFilter(userInfo, &models.ClientEntity{
			OriginClientID: origin.ClientID, ServerID: request.ServerID,
		}, nil)
		if err != nil {
			AbortProblem(c, http.StatusConflict, "FRPS route required", "assign this Server to the node before creating a tunnel")
			return
		}
		if _, err := dao.NewQuery(ctx).GetServerByServerID(userInfo, request.ServerID); err != nil {
			AbortProblem(c, http.StatusNotFound, "Server not found", "the requested FRPS server does not exist")
			return
		}

		base := v1.ProxyBaseConfig{
			Name: request.Name, Type: request.Type,
			ProxyBackend: v1.ProxyBackend{LocalIP: request.LocalHost, LocalPort: request.LocalPort},
		}
		var configurer v1.ProxyConfigurer
		if request.Type == string(v1.ProxyTypeTCP) {
			configurer = &v1.TCPProxyConfig{ProxyBaseConfig: base, RemotePort: request.RemotePort}
		} else {
			configurer = &v1.UDPProxyConfig{ProxyBaseConfig: base, RemotePort: request.RemotePort}
		}
		if err := masterproxy.CreateProxyConfigWithTypedConfig(ctx, masterproxy.CreateProxyConfigWithTypedConfigParam{
			ClientID: request.NodeID, ServerID: request.ServerID,
			ProxyCfg: v1.TypedProxyConfig{Type: request.Type, ProxyConfigurer: configurer}, ClientEntity: child,
		}); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "already exist") {
				AbortProblem(c, http.StatusConflict, "Tunnel already exists", "choose a unique tunnel name")
				return
			}
			AbortProblem(c, http.StatusInternalServerError, "Tunnel creation failed", "could not save or apply the tunnel configuration")
			return
		}
		c.JSON(http.StatusCreated, tunnelResponse{
			Name: request.Name, NodeID: request.NodeID, ClientID: child.ClientID, ServerID: request.ServerID,
			Type: request.Type, LocalHost: request.LocalHost, LocalPort: request.LocalPort, RemotePort: request.RemotePort,
		})
	}
}

func listTunnels(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInfo := common.GetUserInfo(c)
		if userInfo == nil || !userInfo.Valid() {
			AbortProblem(c, http.StatusUnauthorized, "Unauthorized", "a valid user session is required")
			return
		}
		ctx := app.NewContext(c, appInstance)
		configs, err := dao.NewQuery(ctx).ListProxyConfigs(userInfo, 1, 1000)
		if err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Tunnels unavailable", "could not list tunnel configuration")
			return
		}
		tunnels := make([]tunnelResponse, 0, len(configs))
		for _, item := range configs {
			config, err := item.GetTypedProxyConfig()
			if err != nil {
				continue
			}
			base := config.GetBaseConfig()
			remotePort := 0
			switch typed := config.ProxyConfigurer.(type) {
			case *v1.TCPProxyConfig:
				remotePort = typed.RemotePort
			case *v1.UDPProxyConfig:
				remotePort = typed.RemotePort
			default:
				continue
			}
			tunnels = append(tunnels, tunnelResponse{
				Name: item.Name, NodeID: item.OriginClientID, ClientID: item.ClientID, ServerID: item.ServerID,
				Type: item.Type, LocalHost: base.LocalIP, LocalPort: base.LocalPort, RemotePort: remotePort, Stopped: item.Stopped,
			})
		}
		c.JSON(http.StatusOK, gin.H{"tunnels": tunnels})
	}
}

func deleteTunnel(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request deleteTunnelRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			AbortProblem(c, http.StatusBadRequest, "Invalid request", "name, nodeId, and serverId are required")
			return
		}
		userInfo := common.GetUserInfo(c)
		if userInfo == nil || !userInfo.Valid() {
			AbortProblem(c, http.StatusUnauthorized, "Unauthorized", "a valid user session is required")
			return
		}
		ctx := app.NewContext(c, appInstance)
		stored, err := dao.NewQuery(ctx).GetProxyConfigByOriginClientIDAndName(userInfo, request.NodeID, request.Name)
		if err != nil || stored.ServerID != request.ServerID {
			AbortProblem(c, http.StatusNotFound, "Tunnel not found", "the requested tunnel does not exist")
			return
		}
		if _, err := masterproxy.DeleteProxyConfig(ctx, &pb.DeleteProxyConfigRequest{
			ClientId: &stored.ClientID, ServerId: &stored.ServerID, Name: &stored.Name,
		}); err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Tunnel removal failed", "could not remove or apply the tunnel configuration")
			return
		}
		c.Status(http.StatusNoContent)
	}
}

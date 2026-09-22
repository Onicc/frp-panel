package v2

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	masterclient "github.com/Onicc/frp-panel/biz/master/client"
	masterproxy "github.com/Onicc/frp-panel/biz/master/proxy"
	"github.com/Onicc/frp-panel/common"
	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/dao"
	"github.com/Onicc/frp-panel/services/rpc"
	"github.com/Onicc/frp-panel/utils"
	"github.com/Onicc/frp-panel/utils/logger"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/gin-gonic/gin"
)

type createTunnelRequest struct {
	Name       string `json:"name" binding:"required"`
	ClientID   string `json:"clientId" binding:"required"`
	ServerID   string `json:"serverId" binding:"required"`
	Type       string `json:"type" binding:"required"`
	LocalHost  string `json:"localHost"`
	LocalPort  int    `json:"localPort" binding:"required"`
	RemotePort int    `json:"remotePort" binding:"required"`
}

type deleteTunnelRequest struct {
	Name     string `json:"name" binding:"required"`
	ClientID string `json:"clientId" binding:"required"`
	ServerID string `json:"serverId" binding:"required"`
}

type tunnelResponse struct {
	Name       string `json:"name"`
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
			AbortProblem(c, http.StatusBadRequest, "Invalid tunnel", "name, clientId, serverId, type, localPort, and remotePort are required")
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
		origin, err := dao.NewQuery(ctx).GetClientByClientID(userInfo, request.ClientID)
		if err != nil || origin.OriginClientID != "" {
			AbortProblem(c, http.StatusNotFound, "Client not found", "the requested managed Client does not exist")
			return
		}
		server, err := dao.NewQuery(ctx).GetServerByServerID(userInfo, request.ServerID)
		if err != nil {
			AbortProblem(c, http.StatusNotFound, "Server not found", "the requested FRPS server does not exist")
			return
		}
		child, createdConnection, err := ensureClientServerConnection(ctx, origin.ClientEntity, server)
		if err != nil {
			AbortProblem(c, http.StatusConflict, "Server is not ready", "the selected Server must have a valid FRPS configuration")
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
			ClientID: request.ClientID, ServerID: request.ServerID,
			ProxyCfg: v1.TypedProxyConfig{Type: request.Type, ProxyConfigurer: configurer}, ClientEntity: child,
		}); err != nil {
			if createdConnection {
				_ = dao.NewMutation(ctx).DeleteClient(userInfo, child.ClientID)
			}
			if strings.Contains(strings.ToLower(err.Error()), "already exist") {
				AbortProblem(c, http.StatusConflict, "Tunnel already exists", "choose a unique tunnel name")
				return
			}
			AbortProblem(c, http.StatusInternalServerError, "Tunnel creation failed", "could not save or apply the tunnel configuration")
			return
		}
		c.JSON(http.StatusCreated, tunnelResponse{
			Name: request.Name, ClientID: request.ClientID, ServerID: request.ServerID,
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
				Name: item.Name, ClientID: item.OriginClientID, ServerID: item.ServerID,
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
			AbortProblem(c, http.StatusBadRequest, "Invalid request", "name, clientId, and serverId are required")
			return
		}
		userInfo := common.GetUserInfo(c)
		if userInfo == nil || !userInfo.Valid() {
			AbortProblem(c, http.StatusUnauthorized, "Unauthorized", "a valid user session is required")
			return
		}
		ctx := app.NewContext(c, appInstance)
		stored, err := dao.NewQuery(ctx).GetProxyConfigByOriginClientIDAndName(userInfo, request.ClientID, request.Name)
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
		remaining, err := dao.NewQuery(ctx).CountProxyConfigsWithFilters(userInfo, &models.ProxyConfigEntity{ClientID: stored.ClientID})
		if err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Tunnel removal incomplete", "could not inspect the managed FRPC connection")
			return
		}
		if remaining == 0 {
			if err := dao.NewMutation(ctx).DeleteClient(userInfo, stored.ClientID); err != nil {
				AbortProblem(c, http.StatusInternalServerError, "Tunnel removal incomplete", "could not remove the unused managed FRPC connection")
				return
			}
			go notifyClientReconcile(app.NewContext(context.Background(), appInstance), request.ClientID) // #nosec G118 -- application lifecycle owns the RPC connection
		}
		c.Status(http.StatusNoContent)
	}
}

// ensureClientServerConnection creates the internal FRPC connection required by
// a tunnel. It is intentionally not exposed as a separately managed resource:
// users select a Client and Server on each tunnel, and connections are shared.
func ensureClientServerConnection(ctx *app.Context, client *models.ClientEntity, server *models.ServerEntity) (*models.ClientEntity, bool, error) {
	if strings.TrimSpace(server.ServerIP) == "" {
		return nil, false, fmt.Errorf("server %s has no public address", server.ServerID)
	}
	serverConfig, err := server.GetConfigContent()
	if err != nil {
		return nil, false, err
	}
	if serverConfig == nil || serverConfig.BindPort == 0 {
		return nil, false, fmt.Errorf("server %s has no FRPS bind port", server.ServerID)
	}
	connection, created, err := masterclient.ChildClientForServer(ctx, server.ServerID, client)
	if err != nil {
		return nil, false, err
	}
	if !created {
		return connection, false, nil
	}
	userInfo := common.GetUserInfo(ctx)
	config := v1.ClientConfig{ClientCommonConfig: *utils.NewBaseFRPClientUserAuthConfig(
		server.ServerIP, serverConfig.BindPort, utils.FRPClientUser(userInfo.GetUserName(), client.ClientID), userInfo.GetToken(),
	)}
	config.Metadatas[defs.FRPClientIDKey] = connection.ClientID
	if err := connection.SetConfigContent(config); err != nil {
		_ = dao.NewMutation(ctx).DeleteClient(userInfo, connection.ClientID)
		return nil, false, err
	}
	if err := dao.NewMutation(ctx).UpdateClient(userInfo, connection); err != nil {
		_ = dao.NewMutation(ctx).DeleteClient(userInfo, connection.ClientID)
		return nil, false, err
	}
	return connection, true, nil
}

func notifyClientReconcile(ctx *app.Context, clientID string) {
	request := &pb.StartFRPCRequest{ClientId: &clientID}
	if _, err := rpc.CallClient(ctx, clientID, pb.Event_EVENT_START_FRPC, request); err != nil {
		logger.Logger(ctx).WithError(err).Infof("Client %s is offline; connection cleanup will be applied on its next configuration pull", clientID)
	}
}

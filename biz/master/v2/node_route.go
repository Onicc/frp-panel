package v2

import (
	"context"
	"net/http"
	"sort"

	masterclient "github.com/Onicc/frp-panel/biz/master/client"
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

type createNodeRouteRequest struct {
	NodeID   string `json:"nodeId" binding:"required"`
	ServerID string `json:"serverId" binding:"required"`
}

type nodeRoute struct {
	NodeID    string   `json:"nodeId"`
	ServerIDs []string `json:"serverIds"`
}

func createNodeRoute(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request createNodeRouteRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			AbortProblem(c, http.StatusBadRequest, "Invalid request", "nodeId and serverId are required")
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
		server, err := dao.NewQuery(ctx).GetServerByServerID(userInfo, request.ServerID)
		if err != nil {
			AbortProblem(c, http.StatusNotFound, "Server not found", "the requested FRPS server does not exist")
			return
		}
		serverConfig, err := server.GetConfigContent()
		if err != nil || serverConfig.BindPort == 0 {
			AbortProblem(c, http.StatusConflict, "Server is not configured", "configure the FRPS server before assigning this route")
			return
		}

		target, _, err := masterclient.ChildClientForServer(ctx, request.ServerID, origin.ClientEntity)
		if err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Route failed", "could not create the managed FRPC route")
			return
		}
		config := v1.ClientConfig{}
		if len(target.ConfigContent) > 0 {
			current, loadErr := target.GetConfigContent()
			if loadErr != nil {
				AbortProblem(c, http.StatusConflict, "Route failed", "the existing FRPC configuration is invalid")
				return
			}
			config = *current
		}
		config.ClientCommonConfig = *utils.NewBaseFRPClientUserAuthConfig(
			server.ServerIP, serverConfig.BindPort, userInfo.GetUserName(), userInfo.GetToken(),
		)
		config.Metadatas[defs.FRPClientIDKey] = target.ClientID
		if err := target.SetConfigContent(config); err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Route failed", "could not encode the managed FRPC configuration")
			return
		}
		target.ServerID = request.ServerID
		if err := dao.NewMutation(ctx).UpdateClient(userInfo, target); err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Route failed", "could not save the managed FRPC route")
			return
		}

		update := &pb.UpdateFRPCRequest{ClientId: &target.ClientID, ServerId: &target.ServerID, Config: target.ConfigContent}
		go notifyNodeRoute(app.NewContext(context.Background(), appInstance), origin.ClientID, update) // #nosec G118 -- application lifecycle owns the RPC connection
		c.JSON(http.StatusCreated, nodeRoute{NodeID: origin.ClientID, ServerIDs: []string{request.ServerID}})
	}
}

func deleteNodeRoute(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request createNodeRouteRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			AbortProblem(c, http.StatusBadRequest, "Invalid request", "nodeId and serverId are required")
			return
		}
		userInfo := common.GetUserInfo(c)
		if userInfo == nil || !userInfo.Valid() {
			AbortProblem(c, http.StatusUnauthorized, "Unauthorized", "a valid user session is required")
			return
		}
		ctx := app.NewContext(c, appInstance)
		if _, err := dao.NewQuery(ctx).GetClientByClientID(userInfo, request.NodeID); err != nil {
			AbortProblem(c, http.StatusNotFound, "Node not found", "the requested managed node does not exist")
			return
		}
		child, err := dao.NewQuery(ctx).GetClientByFilter(userInfo, &models.ClientEntity{
			OriginClientID: request.NodeID, ServerID: request.ServerID,
		}, nil)
		if err != nil {
			AbortProblem(c, http.StatusNotFound, "Route not found", "the requested FRPC route does not exist")
			return
		}
		mutation := dao.NewMutation(ctx)
		if err := mutation.DeleteProxyConfigsByClientID(userInfo, child.ClientID); err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Route removal failed", "could not remove route proxy configuration")
			return
		}
		if err := mutation.DeleteClient(userInfo, child.ClientID); err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Route removal failed", "could not remove the managed FRPC route")
			return
		}
		go notifyNodeReconcile(app.NewContext(context.Background(), appInstance), request.NodeID) // #nosec G118 -- application lifecycle owns the RPC connection
		c.Status(http.StatusNoContent)
	}
}

func listNodeRoutes(appInstance app.Application) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInfo := common.GetUserInfo(c)
		if userInfo == nil || !userInfo.Valid() {
			AbortProblem(c, http.StatusUnauthorized, "Unauthorized", "a valid user session is required")
			return
		}
		ctx := app.NewContext(c, appInstance)
		nodes, err := dao.NewQuery(ctx).ListClients(userInfo, 1, 1000)
		if err != nil {
			AbortProblem(c, http.StatusInternalServerError, "Routes unavailable", "could not list managed nodes")
			return
		}
		routes := make([]nodeRoute, 0, len(nodes))
		for _, node := range nodes {
			serverIDs := []string{}
			if node.ServerID != "" {
				serverIDs = append(serverIDs, node.ServerID)
			}
			childIDs, err := dao.NewQuery(ctx).GetClientIDsInShadowByClientID(userInfo, node.ClientID)
			if err == nil && len(childIDs) > 0 {
				children, childErr := dao.NewQuery(ctx).GetClientsByClientIDs(userInfo, childIDs)
				if childErr == nil {
					for _, child := range children {
						if child.ServerID != "" {
							serverIDs = append(serverIDs, child.ServerID)
						}
					}
				}
			}
			sort.Strings(serverIDs)
			routes = append(routes, nodeRoute{NodeID: node.ClientID, ServerIDs: uniqueStrings(serverIDs)})
		}
		c.JSON(http.StatusOK, gin.H{"routes": routes})
	}
}

func notifyNodeRoute(ctx *app.Context, originClientID string, update *pb.UpdateFRPCRequest) {
	response, err := rpc.CallClient(ctx, originClientID, pb.Event_EVENT_UPDATE_FRPC, update)
	if err != nil {
		logger.Logger(ctx).WithError(err).Infof("node %s is offline; route will be applied on its next configuration pull", originClientID)
		return
	}
	if response == nil {
		logger.Logger(ctx).Infof("node %s did not acknowledge route update; periodic pull will retry it", originClientID)
	}
}

func notifyNodeReconcile(ctx *app.Context, originClientID string) {
	request := &pb.StartFRPCRequest{ClientId: &originClientID}
	if _, err := rpc.CallClient(ctx, originClientID, pb.Event_EVENT_START_FRPC, request); err != nil {
		logger.Logger(ctx).WithError(err).Infof("node %s is offline; route removal will be applied on its next configuration pull", originClientID)
	}
}

func uniqueStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	result := values[:0]
	for index, value := range values {
		if index == 0 || value != values[index-1] {
			result = append(result, value)
		}
	}
	return result
}

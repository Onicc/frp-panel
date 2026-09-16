package client

import (
	"context"

	"github.com/Onicc/frp-panel/common"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/dao"
	"github.com/Onicc/frp-panel/services/rpc"
	"github.com/Onicc/frp-panel/utils/logger"
)

func DeleteClientHandler(ctx *app.Context, req *pb.DeleteClientRequest) (*pb.DeleteClientResponse, error) {
	logger.Logger(ctx).Infof("delete Client, client: [%s]", req.GetClientId())

	userInfo := common.GetUserInfo(ctx)
	clientID := req.GetClientId()

	if !userInfo.Valid() {
		return &pb.DeleteClientResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "invalid user"},
		}, nil
	}

	if len(clientID) == 0 {
		return &pb.DeleteClientResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "invalid client id"},
		}, nil
	}

	if err := dao.NewMutation(ctx).DeleteClient(userInfo, clientID); err != nil {
		return nil, err
	}

	if err := dao.NewMutation(ctx).DeleteProxyConfigsByClientIDOrOriginClientID(userInfo, clientID); err != nil {
		return nil, err
	}

	go func() {
		resp, err := rpc.CallClient(app.NewContext(context.Background(), ctx.GetApp()), req.GetClientId(), pb.Event_EVENT_REMOVE_FRPC, req)
		if err != nil {
			logger.Logger(context.Background()).WithError(err).Errorf("remove event send to client error, client id: [%s]", req.GetClientId())
		}

		if resp == nil {
			logger.Logger(ctx).Errorf("cannot get response, client id: [%s]", req.GetClientId())
		}
	}()

	return &pb.DeleteClientResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}

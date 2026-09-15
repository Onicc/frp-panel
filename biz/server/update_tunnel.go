package server

import (
	"context"
	"reflect"

	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/server"
	"github.com/Onicc/frp-panel/utils"
	"github.com/Onicc/frp-panel/utils/logger"
)

func UpdateFrpsHander(ctx *app.Context, req *pb.UpdateFRPSRequest) (*pb.UpdateFRPSResponse, error) {
	logger.Logger(ctx).Infof("update frps, req: [%+v]", req)

	content := req.GetConfig()

	s, err := utils.LoadServerConfig(content, true)
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Errorf("cannot load config")
		return &pb.UpdateFRPSResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: err.Error()},
		}, err
	}

	InjectAuthPlugin(ctx, s)

	serverID := req.GetServerId()
	if cli := ctx.GetApp().GetServerController().Get(serverID); cli != nil {
		if reflect.DeepEqual(cli.GetCommonCfg(), s) {
			logger.Logger(ctx).Infof("server %s config not changed", serverID)
			return &pb.UpdateFRPSResponse{
				Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
			}, nil
		}
	}

	handler, err := server.NewServerHandler(s)
	if err != nil {
		logger.Logger(ctx).WithError(err).Error("cannot stage server configuration")
		return &pb.UpdateFRPSResponse{Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: err.Error()}}, err
	}
	if old := ctx.GetApp().GetServerController().Get(serverID); old != nil {
		ctx.GetApp().GetServerController().Delete(serverID)
		logger.Logger(ctx).Infof("server %s config changed, staged replacement is ready", serverID)
	}
	ctx.GetApp().GetServerController().Add(serverID, handler)
	ctx.GetApp().GetServerController().Run(serverID)

	return &pb.UpdateFRPSResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}

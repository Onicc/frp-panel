package server

import (
	"fmt"

	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils/logger"
)

func RemoveFrpsHandler(ctx *app.Context, req *pb.RemoveFRPSRequest) (*pb.RemoveFRPSResponse, error) {
	logger.Logger(ctx).Infof("remove Server runtime, server: [%s]", req.GetServerId())

	if req.GetServerId() == "" {
		return &pb.RemoveFRPSResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "server id cannot be empty"},
		}, fmt.Errorf("server id cannot be empty")
	}

	srv := ctx.GetApp().GetServerController().Get(req.GetServerId())
	if srv == nil {
		logger.Logger(ctx).Infof("server not found, no need to remove")
		return &pb.RemoveFRPSResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "server not found"},
		}, nil
	}

	srv.Stop()
	ctx.GetApp().GetServerController().Delete(req.GetServerId())

	return &pb.RemoveFRPSResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}

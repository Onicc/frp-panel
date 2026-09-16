package client

import (
	"os"
	"time"

	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils/logger"
)

func RemoveFrpcHandler(ctx *app.Context, req *pb.RemoveFRPCRequest) (*pb.RemoveFRPCResponse, error) {
	logger.Logger(ctx).Infof("remove FRPC connection, client: [%s], will exit in 10s", req.GetClientId())

	go func() {
		time.Sleep(10 * time.Second)
		os.Exit(0)
	}()

	return &pb.RemoveFRPCResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}

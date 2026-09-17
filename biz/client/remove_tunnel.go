package client

import (
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils/logger"
)

func RemoveFrpcHandler(ctx *app.Context, req *pb.RemoveFRPCRequest) (*pb.RemoveFRPCResponse, error) {
	clientID := req.GetClientId()
	logger.Logger(ctx).Infof("remove FRPC connections for Client: [%s]", clientID)
	if clientID == "" {
		return &pb.RemoveFRPCResponse{Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "clientId is required"}}, nil
	}
	controller := ctx.GetApp().GetClientController()
	if controller != nil {
		// The legacy wire message carries only clientId. Delete the in-memory
		// FRPC children without terminating the Agent process; newer callers
		// can use the per-server controller API during reconciliation.
		controller.DeleteByClient(clientID)
	}

	return &pb.RemoveFRPCResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}

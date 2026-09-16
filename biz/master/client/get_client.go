package client

import (
	"strings"

	"github.com/Onicc/frp-panel/common"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/dao"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/samber/lo"
)

func GetClientHandler(ctx *app.Context, req *pb.GetClientRequest) (*pb.GetClientResponse, error) {
	logger.Logger(ctx).Infof("get Client, client: [%s], server: [%s]", req.GetClientId(), req.GetServerId())

	var (
		userInfo = common.GetUserInfo(ctx)
		clientID = req.GetClientId()
		serverID = req.GetServerId()
	)

	if !userInfo.Valid() {
		return &pb.GetClientResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "invalid user"},
		}, nil
	}

	if len(clientID) == 0 {
		return &pb.GetClientResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "invalid client id"},
		}, nil
	}

	if !strings.Contains(clientID, ".") {
		clientID = app.GlobalClientID(userInfo.GetUserName(), "c", clientID)
	}

	respCli := &pb.Client{}
	if len(serverID) == 0 {
		client, err := dao.NewQuery(ctx).GetClientByClientID(userInfo, clientID)
		if err != nil {
			return nil, err
		}
		clientIDs, err := dao.NewQuery(ctx).GetClientIDsInShadowByClientID(userInfo, clientID)
		if err != nil {
			logger.Logger(ctx).WithError(err).Errorf("cannot get client ids in shadow, id: [%s]", clientID)
		}

		respCli = &pb.Client{
			Id:        lo.ToPtr(client.ClientID),
			ServerId:  lo.ToPtr(client.ServerID),
			Stopped:   lo.ToPtr(client.Stopped),
			Comment:   lo.ToPtr(client.Comment),
			ClientIds: clientIDs,
			Ephemeral: &client.Ephemeral,
			// LastSeenAt: lo.ToPtr(client.LastSeenAt.UnixMilli()),
		}
	} else {
		client, err := dao.NewQuery(ctx).GetClientByFilter(userInfo, &models.ClientEntity{
			OriginClientID: clientID,
			ServerID:       serverID,
		}, lo.ToPtr(false))
		if err != nil {
			client, err = dao.NewQuery(ctx).GetClientByFilter(userInfo, &models.ClientEntity{
				ClientID: clientID,
				ServerID: serverID,
			}, nil)
			if err != nil {
				return nil, err
			}
		}

		respCli = &pb.Client{
			Id:        lo.ToPtr(client.ClientID),
			ServerId:  lo.ToPtr(client.ServerID),
			Stopped:   lo.ToPtr(client.Stopped),
			Comment:   lo.ToPtr(client.Comment),
			FrpsUrl:   lo.ToPtr(client.FrpsUrl),
			Ephemeral: &client.Ephemeral,
			ClientIds: nil,
		}
		if client.LastSeenAt != nil {
			respCli.LastSeenAt = lo.ToPtr(client.LastSeenAt.UnixMilli())
		}
	}

	return &pb.GetClientResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
		Client: respCli,
	}, nil
}

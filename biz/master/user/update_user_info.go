package user

import (
	"context"
	"net/mail"
	"strings"

	"github.com/Onicc/frp-panel/biz/master/client"
	"github.com/Onicc/frp-panel/common"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/dao"
	"github.com/Onicc/frp-panel/utils"
	"github.com/Onicc/frp-panel/utils/logger"
)

func UpdateUserInfoHander(c *app.Context, req *pb.UpdateUserInfoRequest) (*pb.UpdateUserInfoResponse, error) {
	var (
		userInfo = common.GetUserInfo(c)
	)

	if !userInfo.Valid() {
		return &pb.UpdateUserInfoResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "invalid user"},
		}, nil
	}
	current, ok := userInfo.(*models.UserEntity)
	if !ok {
		return &pb.UpdateUserInfoResponse{Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "invalid user"}}, nil
	}
	newUserEntity := *current
	newUserInfo := req.GetUserInfo()

	if newUserInfo.GetEmail() != "" {
		email := strings.TrimSpace(newUserInfo.GetEmail())
		address, err := mail.ParseAddress(email)
		if err != nil || address.Address != email {
			return &pb.UpdateUserInfoResponse{Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "invalid email"}}, nil
		}
		newUserEntity.Email = email
	}

	if newUserInfo.GetRawPassword() != "" {
		if len(newUserInfo.GetRawPassword()) < 12 {
			return &pb.UpdateUserInfoResponse{Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "password must contain at least 12 characters"}}, nil
		}
		hashedPassword, err := utils.HashPassword(newUserInfo.GetRawPassword())
		if err != nil {
			logger.Logger(context.Background()).WithError(err).Errorf("cannot hash password")
			return nil, err
		}
		newUserEntity.Password = hashedPassword
	}

	if newUserInfo.GetUserName() != "" {
		username := strings.TrimSpace(newUserInfo.GetUserName())
		if !utils.IsClientIDPermited(username) {
			return &pb.UpdateUserInfoResponse{Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "invalid username"}}, nil
		}
		newUserEntity.UserName = username
	}

	if err := dao.NewMutation(c).UpdateUser(userInfo, &newUserEntity); err != nil {
		return &pb.UpdateUserInfoResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: err.Error()},
		}, err
	}

	go func() {
		newUser, err := dao.NewQuery(app.NewContext(context.Background(), c.GetApp())).GetUserByUserID(userInfo.GetUserID())
		if err != nil {
			logger.Logger(context.Background()).WithError(err).Errorf("cannot get user")
			return
		}

		if err := client.SyncTunnel(c, newUser); err != nil {
			logger.Logger(context.Background()).WithError(err).Errorf("cannot sync tunnel, user need to retry update")
		}
	}()

	return &pb.UpdateUserInfoResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}

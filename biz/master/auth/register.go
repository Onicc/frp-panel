package auth

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/dao"
	"github.com/Onicc/frp-panel/utils"
	"github.com/google/uuid"
)

func RegisterHandler(c *app.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	username := req.GetUsername()
	password := req.GetPassword()
	email := req.GetEmail()

	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	address, addressErr := mail.ParseAddress(email)
	if !utils.IsClientIDPermited(username) || len(password) < 12 || addressErr != nil || address.Address != email {
		return &pb.RegisterResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "username, email, or password is invalid"},
		}, fmt.Errorf("username must be URL-safe and password must contain at least 12 characters")
	}

	userCount, err := dao.NewQuery(c).AdminCountUsers()
	if err != nil {
		return &pb.RegisterResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: err.Error()},
		}, err
	}

	if !c.GetApp().GetConfig().App.EnableRegister || userCount > 0 {
		return &pb.RegisterResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: "owner bootstrap is disabled or already complete"},
		}, fmt.Errorf("owner bootstrap is disabled or already complete")
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return &pb.RegisterResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: err.Error()},
		}, err
	}

	newUser := &models.UserEntity{
		UserName: username,
		Password: hashedPassword,
		Email:    email,
		Status:   models.STATUS_NORMAL,
		Role:     defs.UserRole_Viewer,
		Token:    uuid.New().String(),
	}

	if userCount == 0 {
		newUser.Role = defs.UserRole_Owner
	}

	err = dao.NewMutation(c).CreateUser(newUser)
	if err != nil {
		return &pb.RegisterResponse{
			Status: &pb.Status{Code: pb.RespCode_RESP_CODE_INVALID, Message: err.Error()},
		}, err
	}

	return &pb.RegisterResponse{
		Status: &pb.Status{Code: pb.RespCode_RESP_CODE_SUCCESS, Message: "ok"},
	}, nil
}

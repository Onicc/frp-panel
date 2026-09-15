package client

import (
	"github.com/Onicc/frp-panel/biz/common"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
)

func StartPTYConnect(c *app.Context, req *pb.CommonRequest) (*pb.CommonResponse, error) {
	return common.StartPTYConnect(c, req, &pb.PTYClientMessage{Base: &pb.PTYClientMessage_ClientBase{
		ClientBase: &pb.ClientBase{
			ClientId:     c.GetApp().GetConfig().Client.ID,
			ClientSecret: c.GetApp().GetConfig().Client.Secret,
		},
	}})
}

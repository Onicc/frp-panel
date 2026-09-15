package server

import (
	"context"
	"reflect"

	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/server"
	"github.com/Onicc/frp-panel/utils"
	"github.com/Onicc/frp-panel/utils/logger"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/samber/lo"
)

func PullConfig(appInstance app.Application, serverID, serverSecret string) error {
	ctx := app.NewContext(context.Background(), appInstance)

	logger.Logger(ctx).Infof("start to pull server config, serverID: [%s]", serverID)

	cli := appInstance.GetMasterCli()
	resp, err := cli.Call().PullServerConfig(ctx, &pb.PullServerConfigReq{
		Base: &pb.ServerBase{
			ServerId:     serverID,
			ServerSecret: serverSecret,
		},
	})
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Error("cannot pull server config")
		return err
	}

	if len(resp.GetServer().GetConfig()) == 0 {
		logger.Logger(ctx).Infof("server [%s] config is empty, wait for server init", serverID)
		return nil
	}

	s, err := utils.LoadServerConfig([]byte(resp.GetServer().GetConfig()), true)
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Error("cannot load server config")
		return err
	}

	ctrl := appInstance.GetServerController()

	InjectAuthPlugin(ctx, s)

	if t := ctrl.Get(serverID); t != nil {
		if reflect.DeepEqual(t.GetCommonCfg(), s) {
			logger.Logger(ctx).Infof("server %s config not changed", serverID)
			return nil
		}
	}

	handler, err := server.NewServerHandler(s)
	if err != nil {
		logger.Logger(ctx).WithError(err).Error("cannot stage server configuration")
		return err
	}
	if old := ctrl.Get(serverID); old != nil {
		ctrl.Delete(serverID)
		logger.Logger(ctx).Infof("server %s config changed, staged replacement is ready", serverID)
	}
	ctrl.Add(serverID, handler)
	ctrl.Run(serverID)

	logger.Logger(ctx).Infof("pull server config success, serverID: [%s]", serverID)
	return nil
}

func InjectAuthPlugin(ctx *app.Context, cfg *v1.ServerConfig) {
	cfg.HTTPPlugins = lo.Filter(cfg.HTTPPlugins, func(item v1.HTTPPluginOptions, _ int) bool {
		return item.Name != defs.FRP_Plugin_Multiuser
	})
	cfg.HTTPPlugins = append(
		cfg.HTTPPlugins,
		conf.FRPsAuthOption(ctx.GetApp().GetConfig()),
	)
}

package shared

import (
	"context"
	"time"

	bizclient "github.com/Onicc/frp-panel/biz/client"
	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/internal/agent"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/clientrpc"
	"github.com/Onicc/frp-panel/services/rpc"
	"github.com/Onicc/frp-panel/services/tunnel"
	"github.com/Onicc/frp-panel/services/watcher"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
)

type runClientParam struct {
	fx.In

	Lc fx.Lifecycle

	Ctx              *app.Context
	AppInstance      app.Application
	TaskManager      watcher.Client `name:"clientTaskManager"`
	WorkersManager   app.WorkersManager
	WireGuardManager app.WireGuardManager

	Cfg conf.Config
}

func runClient(param runClientParam) {
	var (
		ctx          = param.Ctx
		clientID     = param.AppInstance.GetConfig().Client.ID
		clientSecret = param.AppInstance.GetConfig().Client.Secret
		appInstance  = param.AppInstance
	)
	logger.Logger(ctx).Infof("start to run client")
	if len(clientSecret) == 0 {
		logger.Logger(ctx).Fatal("client secret cannot be empty")
	}

	if len(clientID) == 0 {
		logger.Logger(ctx).Fatal("client id cannot be empty")
	}

	param.TaskManager.AddDurationTask(defs.PullConfigDuration,
		bizclient.PullConfig, appInstance, clientID, clientSecret)
	param.TaskManager.AddDurationTask(defs.PullClientWorkersDuration,
		bizclient.PullWorkers, appInstance, clientID, clientSecret)
	param.TaskManager.AddDurationTask(5*time.Minute, reportClientLocation, context.Background(), appInstance, clientID, clientSecret)
	if appInstance.GetConfig().Client.Features.EnableWireGuard {
		param.TaskManager.AddDurationTask(defs.PullClientWireGuardsDuration,
			bizclient.PullWireGuards, appInstance, clientID, clientSecret)
		param.TaskManager.AddDurationTask(defs.ReportWireGuardRuntimeInfoDuration,
			bizclient.ReportWireGuardRuntimeInfo, appInstance, clientID, clientSecret)
	}

	var wg conc.WaitGroup
	param.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			appInstance.SetRPCCred(NewClientCred(appInstance))
			appInstance.SetMasterCli(rpc.NewMasterCli(appInstance))
			appInstance.SetClientController(tunnel.NewClientController())
			appInstance.SetWireGuardManager(param.WireGuardManager)

			cliRpcHandler := clientrpc.NewClientRPCHandler(
				appInstance,
				clientID,
				clientSecret,
				pb.Event_EVENT_REGISTER_CLIENT,
				bizclient.HandleServerMessage,
			)
			appInstance.SetClientRPCHandler(cliRpcHandler)

			// --- init once start ---
			initClientOnce(appInstance, clientID, clientSecret)
			go reportClientLocation(context.WithoutCancel(ctx), appInstance, clientID, clientSecret)
			initClientWorkerOnce(appInstance, clientID, clientSecret)
			if appInstance.GetConfig().Client.Features.EnableWireGuard {
				initClientWireGuardOnce(appInstance, clientID, clientSecret)
			}
			// --- init once stop ----

			wg.Go(cliRpcHandler.Run)
			wg.Go(param.TaskManager.Run)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			param.TaskManager.Stop()
			appInstance.GetClientRPCHandler().Stop()

			wg.Wait()
			return nil
		},
	})
}

func reportClientLocation(parent context.Context, appInstance app.Application, clientID, clientSecret string) {
	ctx, cancel := context.WithTimeout(parent, 20*time.Second)
	defer cancel()
	ip, err := agent.ProbePublicIP(ctx)
	if err != nil {
		logger.Logger(ctx).WithError(err).Warn("could not probe direct public IP")
		return
	}
	cfg := appInstance.GetConfig()
	if err := agent.ReportPublicIP(ctx, cfg.Client.APIUrl, clientID, clientSecret, ip, cfg.Client.TLSInsecureSkipVerify); err != nil {
		logger.Logger(ctx).WithError(err).Warn("could not report Client location")
	}
}

func initClientOnce(appInstance app.Application, clientID, clientSecret string) {
	err := bizclient.PullConfig(appInstance, clientID, clientSecret)
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Errorf("cannot pull client config, wait for retry")
	}
}

func initClientWorkerOnce(appInstance app.Application, clientID, clientSecret string) {
	err := bizclient.PullWorkers(appInstance, clientID, clientSecret)
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Errorf("cannot pull client workers, wait for retry")
	}
}

func initClientWireGuardOnce(appInstance app.Application, clientID, clientSecret string) {
	err := bizclient.PullWireGuards(appInstance, clientID, clientSecret)
	if err != nil {
		logger.Logger(context.Background()).WithError(err).Errorf("cannot pull client wireguards, wait for retry")
	}
}

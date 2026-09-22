package shared

import (
	"context"
	"time"

	"github.com/Onicc/frp-panel/biz/master/auth"
	"github.com/Onicc/frp-panel/biz/master/proxy"
	v2 "github.com/Onicc/frp-panel/biz/master/v2"
	"github.com/Onicc/frp-panel/biz/master/versions"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/cache"
	"github.com/Onicc/frp-panel/services/master"
	"github.com/Onicc/frp-panel/services/mux"
	"github.com/Onicc/frp-panel/services/watcher"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/Onicc/frp-panel/utils/wsgrpc"
	"github.com/gin-gonic/gin"
	"github.com/sourcegraph/conc"
	"go.uber.org/fx"
)

type runMasterParam struct {
	fx.In

	Lc fx.Lifecycle

	Ctx              *app.Context
	AppInstance      app.Application
	DBManagerMgr     app.DBManager
	HTTPMuxServer    mux.MuxServer `name:"httpMux"`
	TLSMuxServer     mux.MuxServer `name:"tlsMux"`
	MasterRouter     *gin.Engine   `name:"masterRouter"`
	ClientLogManager app.ClientLogManager
	WsGrpcHandler    gin.HandlerFunc      `name:"wsGrpcHandler"`
	MasterService    master.MasterService `name:"wsMasterService"`
	TaskManager      watcher.Client       `name:"masterTaskManager"`
	WsListener       *wsgrpc.WSListener
}

func runMaster(param runMasterParam) {

	param.AppInstance.SetClientLogManager(param.ClientLogManager)
	param.MasterRouter.GET("/wsgrpc", param.WsGrpcHandler)

	cache.InitCache(param.AppInstance.GetConfig())
	auth.InitAuth(param.AppInstance)

	param.TaskManager.AddCronTask("0 0 3 * * *", proxy.CollectDailyStats, param.AppInstance)
	param.TaskManager.AddDurationTask(time.Minute, versions.RefreshAll, param.AppInstance)
	param.TaskManager.AddDurationTask(10*time.Minute, v2.ScheduleServerUpdates, param.AppInstance)

	logger.Logger(param.Ctx).Infof("start to run master")
	var wg conc.WaitGroup

	param.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go versions.RefreshAll(param.AppInstance)       // #nosec G118 -- bounded startup probes
			go v2.ResumeUpdateOperations(param.AppInstance) // #nosec G118 -- resumes bounded update monitoring
			wg.Go(func() {
				if err := param.MasterService.GetServer().Serve(param.WsListener); err != nil {
					logger.Logger(param.Ctx).Fatalf("gRPC server error: %v", err)
				}
			})
			wg.Go(param.TLSMuxServer.Run)
			wg.Go(param.HTTPMuxServer.Run)
			wg.Go(param.TaskManager.Run)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			param.MasterService.GetServer().Stop()
			param.TLSMuxServer.Stop()
			param.HTTPMuxServer.Stop()
			param.TaskManager.Stop()
			wg.Wait()
			return nil
		},
	})
}

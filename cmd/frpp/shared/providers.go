package shared

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	bizcommon "github.com/Onicc/frp-panel/biz/common"
	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/models"
	"github.com/Onicc/frp-panel/pb"
	"github.com/Onicc/frp-panel/services/api"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/services/dao"
	"github.com/Onicc/frp-panel/services/master"
	"github.com/Onicc/frp-panel/services/mux"
	"github.com/Onicc/frp-panel/services/rpc"
	"github.com/Onicc/frp-panel/services/watcher"
	"github.com/Onicc/frp-panel/services/wg"
	"github.com/Onicc/frp-panel/services/workerd"
	"github.com/Onicc/frp-panel/utils"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/Onicc/frp-panel/utils/wsgrpc"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/gorilla/websocket"
	"go.uber.org/fx"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Finish struct {
	fx.Out

	Context context.Context
}

func NewLogHookManager() app.StreamLogHookMgr {
	return &bizcommon.HookMgr{}
}

func NewBaseApp(param struct {
	fx.In

	Cfg     conf.Config `name:"originConfig"`
	CliMgr  app.ClientsManager
	HookMgr app.StreamLogHookMgr
	PtyMgr  app.ShellPTYMgr
}) app.Application {
	appInstance := app.NewApp()
	appInstance.SetConfig(param.Cfg)
	appInstance.SetClientsManager(param.CliMgr)
	appInstance.SetStreamLogHookMgr(param.HookMgr)
	appInstance.SetShellPTYMgr(param.PtyMgr)
	appInstance.SetClientRecvMap(&sync.Map{})
	appInstance.SetNetworkTopologyCache(wg.NewNetworkTopologyCache())
	return appInstance
}

func NewRuntimeConfig(param struct {
	fx.In

	Cfg conf.Config `name:"originConfig"`
}) conf.Config {
	return param.Cfg
}

func NewContext(appInstance app.Application) *app.Context {
	return app.NewContext(context.Background(), appInstance)
}

func NewAndFinishNormalContext(param struct {
	fx.In

	Ctx *app.Context
	Cfg conf.Config
}) Finish {

	return Finish{
		Context: param.Ctx,
	}
}

func NewDBManager(ctx *app.Context, appInstance app.Application) app.DBManager {
	logger.Logger(ctx).Infof("start to init database, type: %s", appInstance.GetConfig().DB.Type)
	mgr := models.NewDBManager(appInstance.GetConfig().DB.Type)
	appInstance.SetDBManager(mgr)

	if appInstance.GetConfig().IsDebug {
		appInstance.GetDBManager().SetDebug(true)
	}

	switch appInstance.GetConfig().DB.Type {
	case defs.DBTypeSQLite3:
		if err := utils.EnsureDirectoryExists(appInstance.GetConfig().DB.DSN); err != nil {
			logger.Logger(ctx).WithError(err).Warnf("ensure directory failed, data location: [%s], keep data in current directory",
				appInstance.GetConfig().DB.DSN)
			tmpCfg := appInstance.GetConfig()
			tmpCfg.DB.DSN = filepath.Base(appInstance.GetConfig().DB.DSN)
			appInstance.SetConfig(tmpCfg)
			logger.Logger(ctx).Infof("new data location: [%s]", appInstance.GetConfig().DB.DSN)
		}

		if sqlitedb, err := gorm.Open(sqlite.Open(appInstance.GetConfig().DB.DSN), databaseConfig()); err != nil {
			logger.Logger(ctx).Panic(err)
		} else {
			appInstance.GetDBManager().SetDB(defs.DBTypeSQLite3, defs.DBRoleDefault, sqlitedb)
			logger.Logger(ctx).Infof("init database success, data location: [%s]", appInstance.GetConfig().DB.DSN)
		}
	case defs.DBTypePostgres:
		if postgresDB, err := gorm.Open(postgres.Open(appInstance.GetConfig().DB.DSN), databaseConfig()); err != nil {
			logger.Logger(ctx).Panic(err)
		} else {
			appInstance.GetDBManager().SetDB(defs.DBTypePostgres, defs.DBRoleDefault, postgresDB)
			logger.Logger(ctx).Infof("init database success, data type: [%s]", "postgres")
		}
	default:
		logger.Logger(ctx).Panicf("currently unsupported database type: %s", appInstance.GetConfig().DB.Type)
	}

	memoryDB, err := gorm.Open(sqlite.Open(":memory:"), databaseConfig())
	if err != nil {
		logger.Logger(ctx).Panic(err)
	}
	appInstance.GetDBManager().SetDB(defs.DBTypeSQLite3, defs.DBRoleRam, memoryDB)
	logger.Logger(ctx).Infof("init memory database success")

	appInstance.GetDBManager().Init()
	return mgr
}

func databaseConfig() *gorm.Config {
	return &gorm.Config{
		// The inherited v1 association graph contains circular ownership edges.
		// v2 migrations create portable tables first; authorization and tenant
		// boundaries are enforced by explicit repository queries.
		DisableForeignKeyConstraintWhenMigrating: true,
		TranslateError:                           true,
	}
}

func NewMasterTLSConfig(ctx *app.Context) *tls.Config {
	return dao.NewMutation(ctx).InitCert(conf.GetCertTemplate(ctx.GetApp().GetConfig()))
}

func NewTLSMasterService(appInstance app.Application, masterTLSConfig *tls.Config) master.MasterService {
	return master.NewMasterService(appInstance, credentials.NewTLS(masterTLSConfig))
}

func NewHTTPMasterService(appInstance app.Application) master.MasterService {
	return master.NewMasterService(appInstance, insecure.NewCredentials())
}

func NewMux(param struct {
	fx.In

	MasterService master.MasterService `name:"tlsMasterService"`
	Router        *gin.Engine          `name:"masterRouter"`
	LisOpt        conf.LisOpt
	TLSCfg        *tls.Config
}) mux.MuxServer {
	return mux.NewMux(param.MasterService.GetServer(), param.Router, param.LisOpt.MuxLis, param.TLSCfg)
}

func NewHTTPMux(param struct {
	fx.In

	MasterService master.MasterService `name:"httpMasterService"`
	Router        *gin.Engine          `name:"masterRouter"`
	LisOpt        conf.LisOpt
}) mux.MuxServer {
	return mux.NewMux(param.MasterService.GetServer(), param.Router, param.LisOpt.ApiLis, nil)
}

func NewWatcher() watcher.Client {
	return watcher.NewClient()
}

func NewWSListener(ctx *app.Context, cfg conf.Config) *wsgrpc.WSListener {
	return wsgrpc.NewWSListener("ws-listener", "wsgrpc", 100)
}

func NewWSGrpcHandler(ctx *app.Context, ws *wsgrpc.WSListener, upgrader *websocket.Upgrader) gin.HandlerFunc {
	return wsgrpc.GinWSHandler(ws, upgrader)
}

func NewWSUpgrader(ctx *app.Context, cfg conf.Config) *websocket.Upgrader {
	return &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return utils.IsOriginAllowed(r, cfg.App.AllowedOrigins) },
	}
}

func NewServerAPI(param struct {
	fx.In
	Ctx          *app.Context
	ServerRouter *gin.Engine `name:"serverRouter"`
}) app.Service {
	l, err := net.Listen("tcp", conf.ServerAPIListenAddr(param.Ctx.GetApp().GetConfig()))
	if err != nil {
		logger.Logger(param.Ctx).WithError(err).Fatalf("failed to listen addr: %v", conf.ServerAPIListenAddr(param.Ctx.GetApp().GetConfig()))
		return nil
	}

	return api.NewApiService(l, param.ServerRouter, true)
}

func NewServerCred(appInstance app.Application) credentials.TransportCredentials {
	cfg := appInstance.GetConfig()
	clientID := cfg.Client.ID
	clientSecret := cfg.Client.Secret
	ctx := context.Background()

	cred, err := utils.TLSClientCert(rpc.GetClientCert(appInstance, clientID, clientSecret, pb.ClientType_CLIENT_TYPE_FRPS))
	if err != nil {
		logger.Logger(ctx).WithError(err).Fatal("new tls client cert failed")
	}
	logger.Logger(ctx).Infof("new tls server cert success")

	return cred
}

func NewClientCred(appInstance app.Application) credentials.TransportCredentials {
	cfg := appInstance.GetConfig()
	clientID := cfg.Client.ID
	clientSecret := cfg.Client.Secret
	ctx := context.Background()

	cred, err := utils.TLSClientCert(rpc.GetClientCert(appInstance, clientID, clientSecret, pb.ClientType_CLIENT_TYPE_FRPC))
	if err != nil {
		logger.Logger(ctx).WithError(err).Fatal("new tls client cert failed")
	}
	logger.Logger(ctx).Infof("new tls client cert success")

	return cred
}

const splitter = "\n--------------------------------------------\n"

func NewRuntimeInfoLogger(param struct {
	fx.In

	Ctx *app.Context
}) {
	logger.Logger(param.Ctx).Info("runtime configuration loaded")
	logger.Logger(param.Ctx).Infof("%scurrent version: \n%s%s", splitter, conf.GetVersion().String(), splitter)
}

func NewWorkersManager(lx fx.Lifecycle, mgr app.WorkerExecManager, appInstance app.Application) app.WorkersManager {
	if !appInstance.GetConfig().Client.Features.EnableFunctions {
		return nil
	}

	workerMgr := workerd.NewWorkersManager()
	appInstance.SetWorkersManager(workerMgr)

	lx.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			workerMgr.StopAllWorkers(app.NewContext(ctx, appInstance))
			logger.Logger(ctx).Info("stop all workers")
			return nil
		},
	})

	return workerMgr
}

func NewWorkerExecManager(cfg conf.Config, appInstance app.Application) app.WorkerExecManager {
	if !appInstance.GetConfig().Client.Features.EnableFunctions {
		return nil
	}

	workerdBinPath := cfg.Client.Worker.WorkerdBinaryPath

	if err := os.MkdirAll(cfg.Client.Worker.WorkerdWorkDir, 0o750); err != nil {
		logger.Logger(context.Background()).WithError(err).Fatalf("create work dir failed, path: [%s]", cfg.Client.Worker.WorkerdWorkDir)
	}

	mgr := workerd.NewExecManager(workerdBinPath,
		[]string{"serve", "--watch", "--verbose"})
	appInstance.SetWorkerExecManager(mgr)
	return mgr
}

func NewWireGuardManager(appInstance app.Application) app.WireGuardManager {
	if !appInstance.GetConfig().Client.Features.EnableWireGuard {
		return nil
	}
	return wg.NewWireGuardManager(appInstance)
}

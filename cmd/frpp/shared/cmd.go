package shared

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Onicc/frp-panel/biz/common/upgrade"
	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/defs"
	serverbootstrap "github.com/Onicc/frp-panel/internal/server"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func BuildCommand(fs embed.FS) *cobra.Command {
	cfg := conf.NewConfig()
	logger.UpdateLoggerOpt(cfg.Logger.FRPLoggerLevel, cfg.Logger.DefaultLoggerLevel, cfg.IsDebug)
	return NewRootCmd(NewMasterCmd(cfg, fs), NewServerCmd(cfg), NewVersionCmd())
}

func NewMasterCmd(cfg conf.Config, fs embed.FS) *cobra.Command {
	return &cobra.Command{
		Use:   "master",
		Short: "Run the frp-panel controller",
		Run: func(_ *cobra.Command, _ []string) {
			if len(strings.TrimSpace(cfg.App.GlobalSecret)) < 32 {
				logger.Logger(context.Background()).Fatal("APP_GLOBAL_SECRET must contain at least 32 non-whitespace characters")
			}
			opts := []fx.Option{
				fx.StartTimeout(defs.AppStartTimeout),
				commonMod,
				masterMod,
				fx.Supply(fx.Annotate(cfg, fx.ResultTags(`name:"originConfig"`)), fs),
				fx.Invoke(NewConfigPrinter),
				fx.Invoke(runMaster),
			}
			if !cfg.IsDebug {
				opts = append(opts, fx.NopLogger)
			}
			controller := fx.New(opts...)
			controller.Run()
			if err := controller.Err(); err != nil {
				logger.Logger(context.Background()).Fatalf("controller application error: %v", err)
			}
		},
	}
}

func NewServerCmd(baseConfig conf.Config) *cobra.Command {
	var configPath, enrollmentToken, apiURL, rpcURL string
	command := &cobra.Command{
		Use:   "server",
		Short: "Run a managed FRPS data-plane server",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg := baseConfig
			persisted, err := serverbootstrap.Resolve(serverbootstrap.ResolveOptions{
				ConfigPath:      configPath,
				EnrollmentToken: enrollmentToken,
				APIURL:          apiURL,
				RPCURL:          rpcURL,
				FallbackID:      cfg.Client.ID,
				FallbackSecret:  cfg.Client.Secret,
				Insecure:        cfg.Client.TLSInsecureSkipVerify,
			})
			if err != nil {
				return err
			}
			cfg.Client.ID = persisted.Credentials.ServerID
			cfg.Client.Secret = persisted.Credentials.Secret
			cfg.Client.APIUrl = persisted.Controller.APIURL
			cfg.Client.RPCUrl = persisted.Controller.RPCURL
			cfg.Client.TLSInsecureSkipVerify = persisted.TLS.InsecureSkipVerify

			opts := []fx.Option{
				fx.StartTimeout(defs.AppStartTimeout),
				commonMod,
				serverMod,
				fx.Supply(fx.Annotate(cfg, fx.ResultTags(`name:"originConfig"`))),
				fx.Invoke(NewConfigPrinter),
				fx.Invoke(runServer),
			}
			if !cfg.IsDebug {
				opts = append(opts, fx.NopLogger)
			}
			server := fx.New(opts...)
			server.Run()
			if err := server.Err(); err != nil {
				return fmt.Errorf("server application error: %w", err)
			}
			return nil
		},
	}
	defaultPath := os.Getenv("SERVER_CONFIG_PATH")
	if defaultPath == "" {
		defaultPath = filepath.Join("/data", "server.yaml")
	}
	command.Flags().StringVar(&configPath, "config", defaultPath, "persistent server configuration path")
	command.Flags().StringVar(&enrollmentToken, "enrollment-token", os.Getenv("SERVER_ENROLLMENT_TOKEN"), "short-lived server enrollment token")
	command.Flags().StringVar(&apiURL, "api-url", baseConfig.Client.APIUrl, "controller public API URL")
	command.Flags().StringVar(&rpcURL, "rpc-url", baseConfig.Client.RPCUrl, "controller public RPC URL")
	return command
}

// RunClientContext runs the managed FRP client until ctx is cancelled. Service
// managers use this entry point so stop and shutdown requests reach fx hooks.
func RunClientContext(ctx context.Context, cfg conf.Config) error {
	opts := []fx.Option{
		fx.StartTimeout(defs.AppStartTimeout),
		clientMod,
		commonMod,
		fx.Supply(fx.Annotate(cfg, fx.ResultTags(`name:"originConfig"`))),
		fx.Invoke(NewConfigPrinter),
		fx.Invoke(runClient),
	}
	if !cfg.IsDebug {
		opts = append(opts, fx.NopLogger)
	}
	agent := fx.New(opts...)
	if err := agent.Err(); err != nil {
		return fmt.Errorf("initialize client: %w", err)
	}
	if err := agent.Start(ctx); err != nil {
		return fmt.Errorf("start client: %w", err)
	}
	<-ctx.Done()
	stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := agent.Stop(stopCtx); err != nil {
		return fmt.Errorf("stop client: %w", err)
	}
	return nil
}

func NewUpgradeWorkerCmd() *cobra.Command {
	worker := &cobra.Command{
		Use:                   "__upgrade-worker",
		Short:                 "Internal upgrade worker",
		Hidden:                true,
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			planPath, _ := cmd.Flags().GetString("plan")
			if planPath == "" {
				return errors.New("missing --plan")
			}
			return upgrade.RunWorker(cmd.Context(), planPath)
		},
	}
	worker.Flags().String("plan", "", "upgrade plan file path")
	_ = worker.Flags().MarkHidden("plan")
	return worker
}

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(conf.GetVersion().String())
		},
	}
}

func NewRootCmd(commands ...*cobra.Command) *cobra.Command {
	root := &cobra.Command{Use: "frp-panel", Short: "Secure FRP control plane"}
	root.AddCommand(commands...)
	return root
}

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Onicc/frp-panel/biz/common/upgrade"
	"github.com/Onicc/frp-panel/cmd/frpp/shared"
	"github.com/Onicc/frp-panel/conf"
	"github.com/Onicc/frp-panel/internal/agent"
	"github.com/Onicc/frp-panel/internal/agentservice"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/fatedier/golib/crypto"
	"github.com/spf13/cobra"
)

func main() {
	crypto.DefaultSalt = "frp"
	logger.InitLogger()
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "frp-panel-agent",
		Short:         "Cross-platform frp-panel node agent",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newAgentCommand(), newServiceCommand(), newDoctorCommand(), newUpdateCommand(), newVersionCommand())
	root.AddCommand(shared.NewUpgradeWorkerCmd())
	return root
}

func newAgentCommand() *cobra.Command {
	agentCommand := &cobra.Command{Use: "agent", Short: "Run the node agent"}
	run := &cobra.Command{
		Use:   "run",
		Short: "Run using a protected YAML configuration file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("config")
			cfg, err := agent.ReadConfig(path)
			if err != nil {
				return err
			}
			caps := agent.CurrentCapabilities()
			if cfg.Features.WireGuard && !caps.WireGuard {
				return fmt.Errorf("wireguard is not supported on %s/%s", caps.OS, caps.Architecture)
			}
			if cfg.Features.Functions && !caps.Functions {
				return fmt.Errorf("functions are not supported on %s/%s", caps.OS, caps.Architecture)
			}
			legacy := cfg.LegacyConfig()
			legacy.Client.Worker.WorkerdWorkDir = filepath.Join(defaultLayout().Data, "workerd")
			return agentservice.RunManaged(func(ctx context.Context) error {
				return shared.RunClientContext(ctx, legacy)
			})
		},
	}
	run.Flags().String("config", defaultLayout().Config, "agent YAML configuration path")
	agentCommand.AddCommand(run)
	return agentCommand
}

func newServiceCommand() *cobra.Command {
	service := &cobra.Command{Use: "service", Short: "Manage the operating-system service"}
	install := &cobra.Command{
		Use:   "install",
		Short: "Install the binary, protected configuration, and service definition",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, _ := cmd.Flags().GetString("root")
			configPath, _ := cmd.Flags().GetString("config")
			start, _ := cmd.Flags().GetBool("start")
			config, err := configForInstall(cmd, configPath)
			if err != nil {
				return err
			}
			encoded, err := agent.EncodeConfig(config)
			if err != nil {
				return err
			}
			layout, err := agentservice.Install(agentservice.InstallOptions{Root: root, Config: encoded, Start: start})
			if err != nil {
				return err
			}
			fmt.Printf("installed %s\nconfig: %s\ndata: %s\n", layout.Binary, layout.Config, layout.Data)
			return nil
		},
	}
	install.Flags().String("root", "", "alternate filesystem root for packaging and safe verification")
	install.Flags().String("config", "", "read an existing agent YAML file")
	install.Flags().String("api-url", "", "controller public API URL")
	install.Flags().String("rpc-url", "", "controller public agent RPC URL")
	install.Flags().String("node-id", "", "node identifier")
	install.Flags().String("enrollment-token", "", "short-lived controller enrollment token")
	install.Flags().String("secret", "", "node credential; stored only in the protected config")
	install.Flags().Bool("insecure-skip-verify", false, "disable TLS verification (unsafe)")
	install.Flags().Bool("enable-functions", false, "enable workerd functions when supported")
	install.Flags().String("workerd-binary", "", "absolute path to an operator-installed workerd binary")
	install.Flags().Bool("enable-remote-shell", false, "enable the privileged remote shell")
	install.Flags().Bool("enable-wireguard", false, "enable WireGuard (Linux only)")
	install.Flags().Bool("start", true, "start or restart the service after installation")

	uninstall := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the service and binary",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root, _ := cmd.Flags().GetString("root")
			purge, _ := cmd.Flags().GetBool("purge")
			_, err := agentservice.Uninstall("", root, purge)
			return err
		},
	}
	uninstall.Flags().String("root", "", "alternate filesystem root")
	uninstall.Flags().Bool("purge", false, "also remove configuration and state")
	service.AddCommand(install, uninstall)
	for _, action := range []string{"status", "start", "stop", "restart"} {
		action := action
		service.AddCommand(&cobra.Command{
			Use:   action,
			Short: action + " the service",
			RunE: func(_ *cobra.Command, _ []string) error {
				return agentservice.Control(action)
			},
		})
	}
	return service
}

func configForInstall(cmd *cobra.Command, path string) (agent.Config, error) {
	if path != "" {
		return agent.ReadConfig(path)
	}
	apiURL, _ := cmd.Flags().GetString("api-url")
	rpcURL, _ := cmd.Flags().GetString("rpc-url")
	nodeID, _ := cmd.Flags().GetString("node-id")
	secret, _ := cmd.Flags().GetString("secret")
	enrollmentToken, _ := cmd.Flags().GetString("enrollment-token")
	insecure, _ := cmd.Flags().GetBool("insecure-skip-verify")
	functions, _ := cmd.Flags().GetBool("enable-functions")
	workerdBinary, _ := cmd.Flags().GetString("workerd-binary")
	remoteShell, _ := cmd.Flags().GetBool("enable-remote-shell")
	wireguard, _ := cmd.Flags().GetBool("enable-wireguard")
	result := agent.Config{
		Version:     agent.ConfigVersion,
		Controller:  agent.Controller{APIURL: apiURL, RPCURL: rpcURL},
		Credentials: agent.Credentials{NodeID: nodeID, Secret: secret},
		TLS:         agent.TLS{InsecureSkipVerify: insecure},
		Features:    agent.Features{Functions: functions, WorkerdBinary: workerdBinary, RemoteShell: remoteShell, WireGuard: wireguard},
	}
	if enrollmentToken != "" {
		joined, err := agent.Enroll(apiURL, enrollmentToken, insecure)
		if err != nil {
			root, _ := cmd.Flags().GetString("root")
			layout, layoutErr := agentservice.LayoutFor(runtime.GOOS, root, os.Getenv)
			if layoutErr == nil {
				if existing, reuseErr := reuseExistingConfig(layout.Config, result); reuseErr == nil {
					fmt.Fprintf(os.Stderr, "enrollment could not be repeated; reusing the matching protected configuration at %s\n", layout.Config)
					return existing, nil
				}
			}
			return agent.Config{}, fmt.Errorf("enroll agent: %w", err)
		}
		if !nodeIDMatchesEnrollment(nodeID, joined.NodeID) {
			return agent.Config{}, fmt.Errorf("enrollment token belongs to a different node")
		}
		result.Credentials.NodeID = joined.NodeID
		result.Credentials.Secret = joined.Secret
	}
	return result, nil
}

func reuseExistingConfig(path string, requested agent.Config) (agent.Config, error) {
	existing, err := agent.ReadConfig(path)
	if err != nil {
		return agent.Config{}, err
	}
	if requested.Credentials.NodeID == "" || !nodeIDMatchesEnrollment(requested.Credentials.NodeID, existing.Credentials.NodeID) {
		return agent.Config{}, errors.New("installed configuration belongs to a different node")
	}
	if !sameEndpoint(requested.Controller.APIURL, existing.Controller.APIURL) ||
		!sameEndpoint(requested.Controller.RPCURL, existing.Controller.RPCURL) {
		return agent.Config{}, errors.New("installed configuration belongs to a different controller")
	}
	return existing, nil
}

func sameEndpoint(left, right string) bool {
	return left != "" && strings.TrimRight(left, "/") == strings.TrimRight(right, "/")
}

func nodeIDMatchesEnrollment(requested, enrolled string) bool {
	return requested == "" || enrolled == requested || strings.HasSuffix(enrolled, ".c."+requested)
}

func newDoctorCommand() *cobra.Command {
	var configPath string
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Inspect platform capabilities and installation health",
		RunE: func(_ *cobra.Command, _ []string) error {
			layout := defaultLayout()
			result := struct {
				Capabilities agent.Capabilities `json:"capabilities"`
				Binary       string             `json:"binary"`
				Config       string             `json:"config"`
				Data         string             `json:"data"`
				ConfigValid  bool               `json:"configValid"`
				ConfigError  string             `json:"configError,omitempty"`
			}{Capabilities: agent.CurrentCapabilities(), Binary: layout.Binary, Config: configPath, Data: layout.Data}
			if _, err := agent.ReadConfig(configPath); err != nil {
				result.ConfigError = err.Error()
			} else {
				result.ConfigValid = true
			}
			if asJSON {
				return json.NewEncoder(os.Stdout).Encode(result)
			}
			fmt.Printf("platform: %s/%s\nbinary: %s\nconfig: %s\ndata: %s\nconfig valid: %t\n", result.Capabilities.OS, result.Capabilities.Architecture, result.Binary, result.Config, result.Data, result.ConfigValid)
			if result.ConfigError != "" {
				fmt.Printf("config error: %s\n", result.ConfigError)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&configPath, "config", defaultLayout().Config, "agent YAML configuration path")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print machine-readable JSON")
	return cmd
}

func newUpdateCommand() *cobra.Command {
	var version string
	var restart bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Install a verified release with backup and rollback support",
		RunE: func(command *cobra.Command, _ []string) error {
			_, err := upgrade.StartWithResult(command.Context(), upgrade.Options{
				Version: version, Backup: true, RestartService: restart,
				ServiceName: agentservice.ServiceName,
			})
			return err
		},
	}
	cmd.Flags().StringVar(&version, "version", "edge", "release tag or latest stable version")
	cmd.Flags().BoolVar(&restart, "restart-service", true, "restart service after a successful replacement")
	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version and platform",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(conf.GetVersion().String())
		},
	}
}

func defaultLayout() agentservice.Layout {
	layout, err := agentservice.LayoutFor(runtime.GOOS, "", os.Getenv)
	if err != nil {
		panic(err)
	}
	return layout
}

package agent

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Onicc/frp-panel/conf"
	"gopkg.in/yaml.v3"
)

const ConfigVersion = 2

// Config is the on-disk agent configuration. Credentials intentionally live in
// a mode-0600 file instead of the service command line.
type Config struct {
	Version     int         `yaml:"version"`
	Controller  Controller  `yaml:"controller"`
	Credentials Credentials `yaml:"credentials"`
	TLS         TLS         `yaml:"tls"`
	Features    Features    `yaml:"features"`
}

type Controller struct {
	APIURL string `yaml:"api_url"`
	RPCURL string `yaml:"rpc_url"`
}

type Credentials struct {
	NodeID string `yaml:"node_id"`
	Secret string `yaml:"secret"`
}

type TLS struct {
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

type Features struct {
	Functions     bool   `yaml:"functions"`
	WorkerdBinary string `yaml:"workerd_binary,omitempty"`
	RemoteShell   bool   `yaml:"remote_shell"`
	WireGuard     bool   `yaml:"wireguard"`
}

func (c Config) Validate() error {
	if c.Version != ConfigVersion {
		return fmt.Errorf("unsupported config version %d", c.Version)
	}
	if c.Controller.APIURL == "" || c.Controller.RPCURL == "" {
		return errors.New("controller.api_url and controller.rpc_url are required")
	}
	if c.Credentials.NodeID == "" || c.Credentials.Secret == "" {
		return errors.New("credentials.node_id and credentials.secret are required")
	}
	if c.Features.Functions && (c.Features.WorkerdBinary == "" || !filepath.IsAbs(c.Features.WorkerdBinary)) {
		return errors.New("features.workerd_binary must be an absolute path when functions are enabled")
	}
	return nil
}

func ReadConfig(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read agent config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode agent config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func EncodeConfig(cfg Config) ([]byte, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return yaml.Marshal(cfg)
}

func (c Config) LegacyConfig() conf.Config {
	cfg := conf.DefaultConfig()
	cfg.Client.ID = c.Credentials.NodeID
	cfg.Client.Secret = c.Credentials.Secret
	cfg.Client.APIUrl = c.Controller.APIURL
	cfg.Client.RPCUrl = c.Controller.RPCURL
	cfg.Client.TLSInsecureSkipVerify = c.TLS.InsecureSkipVerify
	cfg.Client.Features.EnableFunctions = c.Features.Functions
	cfg.Client.Features.EnableRemoteShell = c.Features.RemoteShell
	cfg.Client.Features.EnableWireGuard = c.Features.WireGuard
	cfg.Client.Worker.WorkerdBinaryPath = c.Features.WorkerdBinary
	return cfg
}

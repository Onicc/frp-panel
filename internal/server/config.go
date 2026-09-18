package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ConfigVersion = 2

// Config is persisted on the FRPS host after a one-time enrollment. Keeping
// the permanent credential in a mode-0600 volume file avoids exposing it in a
// Compose file, process arguments, or the Master database.
type Config struct {
	Version             int         `yaml:"version"`
	Master              Master      `yaml:"master"`
	Credentials         Credentials `yaml:"credentials"`
	TLS                 TLS         `yaml:"tls"`
	EnrollmentTokenHash string      `yaml:"enrollment_token_hash,omitempty"`
}

type Master struct {
	APIURL string `yaml:"api_url"`
	RPCURL string `yaml:"rpc_url"`
}

type Credentials struct {
	ServerID string `yaml:"server_id"`
	Secret   string `yaml:"secret"`
}

type TLS struct {
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

func (c Config) Validate() error {
	if c.Version != ConfigVersion {
		return fmt.Errorf("unsupported server config version %d", c.Version)
	}
	if c.Master.APIURL == "" || c.Master.RPCURL == "" {
		return errors.New("master.api_url and master.rpc_url are required")
	}
	if c.Credentials.ServerID == "" || c.Credentials.Secret == "" {
		return errors.New("credentials.server_id and credentials.secret are required")
	}
	return nil
}

func ReadConfig(path string) (Config, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- the operator explicitly selects the persistent config path
	if err != nil {
		return Config{}, fmt.Errorf("read server config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode server config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func WriteConfig(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	raw, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode server config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create server config directory: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".server-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary server config: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("protect temporary server config: %w", err)
	}
	if _, err := temporary.Write(raw); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary server config: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync temporary server config: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary server config: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("install server config: %w", err)
	}
	return nil
}

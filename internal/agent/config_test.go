package agent

import (
	"path/filepath"
	"testing"
)

func validConfig() Config {
	return Config{
		Version: ConfigVersion,
		Master: Master{
			APIURL: "https://panel.example.test",
			RPCURL: "wss://panel.example.test",
		},
		Credentials: Credentials{ClientID: "owner.c.client", Secret: "secret"},
	}
}

func TestConfigRequiresExplicitWorkerdPath(t *testing.T) {
	cfg := validConfig()
	cfg.Features.Functions = true
	if err := cfg.Validate(); err == nil {
		t.Fatal("functions accepted without an explicit workerd binary")
	}
	cfg.Features.WorkerdBinary = filepath.Join(string(filepath.Separator), "opt", "workerd")
	if err := cfg.Validate(); err != nil {
		t.Fatalf("absolute workerd path rejected: %v", err)
	}
}

func TestLegacyConfigIgnoresProcessClientEnvironment(t *testing.T) {
	t.Setenv("CLIENT_SECRET", "environment-secret")
	cfg := validConfig()
	cfg.Features.WireGuard = true
	legacy := cfg.LegacyConfig()
	if legacy.Client.Secret != "secret" || !legacy.Client.Features.EnableWireGuard {
		t.Fatalf("unexpected runtime config: %+v", legacy.Client)
	}
}

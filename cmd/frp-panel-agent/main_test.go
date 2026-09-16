package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Onicc/frp-panel/internal/agent"
	"github.com/Onicc/frp-panel/internal/agentservice"
)

func TestNodeIDMatchesEnrollment(t *testing.T) {
	tests := []struct {
		requested, enrolled string
		want                bool
	}{
		{"owner.c.mac", "owner.c.mac", true},
		{"mac", "owner.c.mac", true},
		{"", "owner.c.mac", true},
		{"other", "owner.c.mac", false},
	}
	for _, test := range tests {
		if got := nodeIDMatchesEnrollment(test.requested, test.enrolled); got != test.want {
			t.Fatalf("nodeIDMatchesEnrollment(%q, %q) = %t, want %t", test.requested, test.enrolled, got, test.want)
		}
	}
}

func TestReuseExistingConfigRequiresMatchingNodeAndController(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.yaml")
	existing := agent.Config{
		Version:     agent.ConfigVersion,
		Controller:  agent.Controller{APIURL: "https://panel.example.com", RPCURL: "wss://panel.example.com"},
		Credentials: agent.Credentials{NodeID: "owner.c.node", Secret: "stored-secret"},
	}
	raw, err := agent.EncodeConfig(existing)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	requested := agent.Config{
		Controller:  agent.Controller{APIURL: "https://panel.example.com/", RPCURL: "wss://panel.example.com/"},
		Credentials: agent.Credentials{NodeID: "node"},
	}
	got, err := reuseExistingConfig(path, requested)
	if err != nil {
		t.Fatal(err)
	}
	if got.Credentials.Secret != existing.Credentials.Secret {
		t.Fatalf("reused secret = %q, want stored credential", got.Credentials.Secret)
	}

	requested.Credentials.NodeID = "different"
	if _, err := reuseExistingConfig(path, requested); err == nil {
		t.Fatal("expected a node mismatch to reject the installed configuration")
	}
	requested.Credentials.NodeID = ""
	if _, err := reuseExistingConfig(path, requested); err == nil {
		t.Fatal("expected an omitted node ID to reject the installed configuration")
	}
	requested.Credentials.NodeID = "node"
	requested.Controller.APIURL = "https://other.example.com"
	if _, err := reuseExistingConfig(path, requested); err == nil {
		t.Fatal("expected a controller mismatch to reject the installed configuration")
	}
}

func TestConfigForInstallReusesMatchingConfigAfterEnrollmentFailure(t *testing.T) {
	controller := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "expired", http.StatusUnauthorized)
	}))
	defer controller.Close()

	root := t.TempDir()
	layout, err := agentservice.LayoutFor(runtime.GOOS, root, os.Getenv)
	if err != nil {
		t.Fatal(err)
	}
	existing := agent.Config{
		Version:     agent.ConfigVersion,
		Controller:  agent.Controller{APIURL: controller.URL, RPCURL: "ws://controller.example.test"},
		Credentials: agent.Credentials{NodeID: "owner.c.node", Secret: "stored-secret"},
	}
	raw, err := agent.EncodeConfig(existing)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(layout.Config), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(layout.Config, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	service := newServiceCommand()
	install, _, err := service.Find([]string{"install"})
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{
		"root":             root,
		"api-url":          controller.URL,
		"rpc-url":          "ws://controller.example.test",
		"node-id":          "node",
		"enrollment-token": "already-consumed",
	} {
		if err := install.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}

	got, err := configForInstall(install, "")
	if err != nil {
		t.Fatal(err)
	}
	if got.Credentials.Secret != existing.Credentials.Secret {
		t.Fatalf("reused secret = %q, want stored credential", got.Credentials.Secret)
	}
}

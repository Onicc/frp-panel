package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Onicc/frp-panel/biz/common/upgrade"
	"github.com/Onicc/frp-panel/internal/agent"
	"github.com/Onicc/frp-panel/internal/agentservice"
)

func TestAgentUpdateRestartModesWithoutTouchingLocalService(t *testing.T) {
	for _, tc := range []struct {
		goos           string
		restart        bool
		nativeCalls    int
		libraryRestart bool
	}{
		{"darwin", true, 1, false},
		{"darwin", false, 0, false},
		{"linux", true, 0, true},
	} {
		calls := 0
		start := func(_ context.Context, options upgrade.Options) error {
			if options.Version != "edge" || options.RestartService != tc.libraryRestart || !options.Backup {
				t.Fatalf("unexpected options: %+v", options)
			}
			return nil
		}
		control := func(action string) error {
			if action != "restart" {
				t.Fatalf("unexpected action: %q", action)
			}
			calls++
			return nil
		}
		if err := runAgentUpdate(context.Background(), "edge", tc.restart, tc.goos, start, control); err != nil {
			t.Fatal(err)
		}
		if calls != tc.nativeCalls {
			t.Fatalf("%s: native calls = %d, want %d", tc.goos, calls, tc.nativeCalls)
		}
	}
	controlCalled := false
	err := runAgentUpdate(context.Background(), "edge", true, "darwin", func(context.Context, upgrade.Options) error { return errors.New("download failed") }, func(string) error { controlCalled = true; return nil })
	if err == nil || controlCalled {
		t.Fatal("failed download must never restart the service")
	}
}

func TestClientIDMatchesEnrollment(t *testing.T) {
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
		if got := clientIDMatchesEnrollment(test.requested, test.enrolled); got != test.want {
			t.Fatalf("clientIDMatchesEnrollment(%q, %q) = %t, want %t", test.requested, test.enrolled, got, test.want)
		}
	}
}

func TestReuseExistingConfigRequiresMatchingClientAndController(t *testing.T) {
	path := filepath.Join(t.TempDir(), "agent.yaml")
	existing := agent.Config{
		Version:     agent.ConfigVersion,
		Master:      agent.Master{APIURL: "https://panel.example.com", RPCURL: "wss://panel.example.com"},
		Credentials: agent.Credentials{ClientID: "owner.c.client", Secret: "stored-secret"},
	}
	raw, err := agent.EncodeConfig(existing)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	requested := agent.Config{
		Master:      agent.Master{APIURL: "https://panel.example.com/", RPCURL: "wss://panel.example.com/"},
		Credentials: agent.Credentials{ClientID: "client"},
	}
	got, err := reuseExistingConfig(path, requested)
	if err != nil {
		t.Fatal(err)
	}
	if got.Credentials.Secret != existing.Credentials.Secret {
		t.Fatalf("reused secret = %q, want stored credential", got.Credentials.Secret)
	}

	requested.Credentials.ClientID = "different"
	if _, err := reuseExistingConfig(path, requested); err == nil {
		t.Fatal("expected a Client mismatch to reject the installed configuration")
	}
	requested.Credentials.ClientID = ""
	if _, err := reuseExistingConfig(path, requested); err == nil {
		t.Fatal("expected an omitted Client ID to reject the installed configuration")
	}
	requested.Credentials.ClientID = "client"
	requested.Master.APIURL = "https://other.example.com"
	if _, err := reuseExistingConfig(path, requested); err == nil {
		t.Fatal("expected a Master mismatch to reject the installed configuration")
	}
}

func TestConfigForInstallReusesMatchingConfigAfterEnrollmentFailure(t *testing.T) {
	masterServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "expired", http.StatusUnauthorized)
	}))
	defer masterServer.Close()

	root := t.TempDir()
	layout, err := agentservice.LayoutFor(runtime.GOOS, root, os.Getenv)
	if err != nil {
		t.Fatal(err)
	}
	existing := agent.Config{
		Version:     agent.ConfigVersion,
		Master:      agent.Master{APIURL: masterServer.URL, RPCURL: "ws://master.example.test"},
		Credentials: agent.Credentials{ClientID: "owner.c.client", Secret: "stored-secret"},
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
		"api-url":          masterServer.URL,
		"rpc-url":          "ws://master.example.test",
		"client-id":        "client",
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

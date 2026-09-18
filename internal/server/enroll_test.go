package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/Onicc/frp-panel/utils"
)

func TestResolveEnrollsOnceThenUsesPersistedCredentials(t *testing.T) {
	requests := 0
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/server/enroll" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(EnrollmentResult{ServerID: "owner.s.edge", Secret: "permanent-secret"})
	}))
	defer httpServer.Close()

	path := filepath.Join(t.TempDir(), "server.yaml")
	options := ResolveOptions{ConfigPath: path, EnrollmentToken: "one-use-token", APIURL: httpServer.URL, RPCURL: "ws://master.test"}
	first, err := Resolve(options)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Resolve(options)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 1 || first != second || second.Credentials.ServerID != "owner.s.edge" {
		t.Fatalf("unexpected resolve result: requests=%d first=%#v second=%#v", requests, first, second)
	}
}

func TestResolveReenrollsWhenTokenChanges(t *testing.T) {
	var tokens []string
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode enrollment request: %v", err)
		}
		tokens = append(tokens, request.Token)
		_ = json.NewEncoder(w).Encode(EnrollmentResult{ServerID: "owner.s.edge", Secret: "secret-" + request.Token})
	}))
	defer httpServer.Close()

	path := filepath.Join(t.TempDir(), "server.yaml")
	first, err := Resolve(ResolveOptions{ConfigPath: path, EnrollmentToken: "first-token", APIURL: httpServer.URL, RPCURL: "ws://master.test"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := Resolve(ResolveOptions{ConfigPath: path, EnrollmentToken: "second-token", APIURL: httpServer.URL, RPCURL: "ws://master.test"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 || tokens[0] != "first-token" || tokens[1] != "second-token" {
		t.Fatalf("unexpected enrollment requests: %#v", tokens)
	}
	if first.Credentials.Secret == second.Credentials.Secret || second.Credentials.Secret != "secret-second-token" {
		t.Fatalf("credentials were not refreshed: first=%#v second=%#v", first.Credentials, second.Credentials)
	}
	if !utils.CheckCredential("second-token", second.EnrollmentTokenHash) {
		t.Fatalf("new enrollment token fingerprint was not persisted")
	}
}

func TestResolveKeepsLegacyConfigWhenTokenWasAlreadyConsumed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "server.yaml")
	persisted := Config{
		Version:     ConfigVersion,
		Master:      Master{APIURL: "https://panel.example.test", RPCURL: "wss://panel.example.test"},
		Credentials: Credentials{ServerID: "owner.s.edge", Secret: "persisted-secret"},
	}
	if err := WriteConfig(path, persisted); err != nil {
		t.Fatal(err)
	}
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"detail":"the token is already used"}`))
	}))
	defer httpServer.Close()

	actual, err := Resolve(ResolveOptions{ConfigPath: path, EnrollmentToken: "already-consumed", APIURL: httpServer.URL, RPCURL: "ws://master.test"})
	if err != nil {
		t.Fatal(err)
	}
	if actual != persisted {
		t.Fatalf("legacy config changed after failed migration: %#v", actual)
	}
}

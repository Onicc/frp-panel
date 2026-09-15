package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
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
	options := ResolveOptions{ConfigPath: path, EnrollmentToken: "one-use-token", APIURL: httpServer.URL, RPCURL: "ws://controller.test"}
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

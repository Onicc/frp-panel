package release

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Onicc/frp-panel/conf"
)

func TestCheckerStableAndLegacyChannels(t *testing.T) {
	oldCommit := strings.Repeat("a", 40)
	newCommit := strings.Repeat("b", 40)
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/releases/latest":
			fmt.Fprintf(w, `{"id":2,"tag_name":"v1.2.0","target_commitish":%q,"assets":[{"id":1,"name":"frp-panel-linux-amd64"}]}`, newCommit)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	checker := NewChecker()
	checker.BaseURL = server.URL
	for _, tc := range []struct {
		name, version, commit, channel string
		available                      *bool
	}{
		{"edge legacy", "edge-SNAPSHOT", oldCommit, "legacy", nil},
		{"main legacy", "main", newCommit, "legacy", nil},
		{"stable older", "v1.1.0", oldCommit, "stable", boolPointer(true)},
		{"stable same", "v1.2.0", newCommit, "stable", boolPointer(false)},
		{"stable newer", "v1.3.0", oldCommit, "stable", boolPointer(false)},
		{"development", "dev", oldCommit, "unknown", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			snapshot, _ := checker.Check(context.Background(), conf.VersionInfo{GitVersion: tc.version, GitCommit: tc.commit}, false)
			if snapshot.Channel != tc.channel || (snapshot.Available == nil) != (tc.available == nil) {
				t.Fatalf("unexpected result: %+v", snapshot)
			}
			if tc.available != nil && *snapshot.Available != *tc.available {
				t.Fatalf("unexpected availability: %+v", snapshot)
			}
		})
	}
	_, _ = checker.Check(context.Background(), conf.VersionInfo{GitVersion: "edge-SNAPSHOT", GitCommit: oldCommit}, false)
	if requestCount != 1 {
		t.Fatalf("expected cached release metadata; got %d HTTP requests", requestCount)
	}
}

func TestCheckerRejectsReusedOrPrereleaseTags(t *testing.T) {
	oldCommit := strings.Repeat("a", 40)
	newCommit := strings.Repeat("b", 40)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":2,"tag_name":"v1.2.0","target_commitish":%q,"assets":[{"id":1,"name":"frp-panel-linux-amd64"}]}`, newCommit)
	}))
	defer server.Close()
	checker := NewChecker()
	checker.BaseURL = server.URL
	snapshot, _ := checker.Check(context.Background(), conf.VersionInfo{GitVersion: "v1.2.0", GitCommit: oldCommit}, false)
	if snapshot.Available != nil || !strings.Contains(snapshot.Error, "different commit") {
		t.Fatalf("reused tag must not be updatable: %+v", snapshot)
	}
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":2,"tag_name":"v1.3.0","draft":true,"target_commitish":%q,"assets":[{"id":1,"name":"frp-panel-linux-amd64"}]}`, newCommit)
	})
	checker = NewChecker()
	checker.BaseURL = server.URL
	if _, err := checker.Latest(context.Background(), "stable", false); err == nil {
		t.Fatal("draft release must be rejected")
	}
}

func boolPointer(value bool) *bool { return &value }

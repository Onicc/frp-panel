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

func TestCheckerChannelsAndCommitAncestry(t *testing.T) {
	oldCommit := strings.Repeat("a", 40)
	newCommit := strings.Repeat("b", 40)
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/releases/tags/edge":
			fmt.Fprintf(w, `{"id":1,"tag_name":"edge","target_commitish":%q,"assets":[{"id":1,"name":"frp-panel-linux-amd64"}]}`, newCommit)
		case r.URL.Path == "/releases/latest":
			fmt.Fprintf(w, `{"id":2,"tag_name":"v1.2.0","target_commitish":%q,"assets":[{"id":1,"name":"frp-panel-linux-amd64"}]}`, newCommit)
		case r.URL.Path == "/compare/"+oldCommit+"..."+newCommit:
			fmt.Fprint(w, `{"status":"ahead"}`)
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
		{"edge older", "edge-SNAPSHOT", oldCommit, "edge", boolPointer(true)},
		{"edge same", "main", newCommit, "edge", boolPointer(false)},
		{"stable older", "v1.1.0", oldCommit, "stable", boolPointer(true)},
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
	if requestCount != 3 { // one fetch per channel and one ancestry check
		t.Fatalf("expected cached release metadata; got %d HTTP requests", requestCount)
	}
}

func boolPointer(value bool) *bool { return &value }

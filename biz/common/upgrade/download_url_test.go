package upgrade

import (
	"strings"
	"testing"
)

func TestBuildDownloadURLUsesStableReleases(t *testing.T) {
	for _, tc := range []struct {
		version string
		part    string
	}{
		{"", "/releases/latest/download/"},
		{"latest", "/releases/latest/download/"},
		{"v2.1.0", "/releases/download/v2.1.0/"},
	} {
		url, err := buildDownloadURL(Options{Version: tc.version})
		if err != nil || !strings.Contains(url, tc.part) {
			t.Fatalf("version %q: url=%q, err=%v", tc.version, url, err)
		}
	}
	for _, invalid := range []string{"edge", "edge-SNAPSHOT", "main", "v2.1.0-rc1", "../../other"} {
		if _, err := buildDownloadURL(Options{Version: invalid}); err == nil {
			t.Fatalf("version %q must not be accepted", invalid)
		}
	}
}

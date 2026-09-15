package agent

import "testing"

func TestDetectCapabilities(t *testing.T) {
	tests := []struct {
		name      string
		goos      string
		goarch    string
		functions bool
		wireguard bool
	}{
		{name: "linux amd64", goos: "linux", goarch: "amd64", functions: true, wireguard: true},
		{name: "mac arm64", goos: "darwin", goarch: "arm64", functions: true, wireguard: false},
		{name: "windows amd64", goos: "windows", goarch: "amd64", functions: false, wireguard: false},
		{name: "windows arm64", goos: "windows", goarch: "arm64", functions: false, wireguard: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectCapabilities(tt.goos, tt.goarch)
			if !got.FRPC || got.FRPS || got.Functions != tt.functions || got.WireGuard != tt.wireguard {
				t.Fatalf("unexpected capabilities: %+v", got)
			}
		})
	}
}

//go:build !windows

package upgrade

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func detectAssetName() (string, error) {
	osName := runtime.GOOS
	binary := "frp-panel"
	if executable, err := os.Executable(); err == nil && strings.Contains(filepath.Base(executable), "agent") {
		binary = "frp-panel-agent"
	}
	machine := unameMachine()
	if len(machine) == 0 {
		// fallback
		machine = runtime.GOARCH
	}

	switch osName {
	case "linux":
		switch machine {
		case "x86_64", "amd64":
			return binary + "-linux-amd64", nil
		case "aarch64", "arm64":
			return binary + "-linux-arm64", nil
		}
	case "darwin":
		switch machine {
		case "x86_64", "amd64":
			return binary + "-darwin-amd64", nil
		case "arm64":
			return binary + "-darwin-arm64", nil
		}
	}
	return "", fmt.Errorf("暂不支持的系统/架构: %s %s", osName, machine)
}

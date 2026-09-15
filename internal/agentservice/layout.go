package agentservice

import (
	"fmt"
	"path/filepath"
	"strings"
)

const ServiceName = "frp-panel-agent"

type Layout struct {
	GOOS        string
	Binary      string
	Config      string
	Data        string
	ServiceFile string
	ServiceUser string
}

func LayoutFor(goos, root string, getenv func(string) string) (Layout, error) {
	if root == "" {
		root = string(filepath.Separator)
	}
	rebase := func(path string) string {
		if root == string(filepath.Separator) {
			return path
		}
		volume := filepath.VolumeName(path)
		path = strings.TrimPrefix(path, volume)
		path = strings.TrimLeft(path, "/\\")
		return filepath.Join(root, filepath.FromSlash(strings.ReplaceAll(path, `\\`, "/")))
	}

	switch goos {
	case "linux":
		return Layout{
			GOOS:        goos,
			Binary:      rebase("/usr/local/libexec/frp-panel/frp-panel-agent"),
			Config:      rebase("/etc/frp-panel/agent.yaml"),
			Data:        rebase("/var/lib/frp-panel"),
			ServiceFile: rebase("/etc/systemd/system/frp-panel-agent.service"),
			ServiceUser: "frp-panel",
		}, nil
	case "darwin":
		return Layout{
			GOOS:        goos,
			Binary:      rebase("/usr/local/libexec/frp-panel/frp-panel-agent"),
			Config:      rebase("/Library/Application Support/frp-panel/agent.yaml"),
			Data:        rebase("/Library/Application Support/frp-panel"),
			ServiceFile: rebase("/Library/LaunchDaemons/io.github.onicc.frp-panel.agent.plist"),
			ServiceUser: "_frp-panel",
		}, nil
	case "windows":
		programFiles := getenv("ProgramFiles")
		programData := getenv("ProgramData")
		if programFiles == "" {
			programFiles = `C:\Program Files`
		}
		if programData == "" {
			programData = `C:\ProgramData`
		}
		return Layout{
			GOOS:        goos,
			Binary:      rebase(filepath.Join(programFiles, "frp-panel", "frp-panel-agent.exe")),
			Config:      rebase(filepath.Join(programData, "frp-panel", "agent.yaml")),
			Data:        rebase(filepath.Join(programData, "frp-panel")),
			ServiceUser: `NT AUTHORITY\LocalService`,
		}, nil
	default:
		return Layout{}, fmt.Errorf("unsupported operating system %q", goos)
	}
}

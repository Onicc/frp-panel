package agent

import "runtime"

type Capabilities struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	FRPC         bool   `json:"frpc"`
	FRPS         bool   `json:"frps"`
	Functions    bool   `json:"functions"`
	RemoteShell  bool   `json:"remoteShell"`
	WireGuard    bool   `json:"wireGuard"`
}

func DetectCapabilities(goos, goarch string) Capabilities {
	return Capabilities{
		OS:           goos,
		Architecture: goarch,
		FRPC:         true,
		FRPS:         false,
		Functions:    goos != "windows",
		RemoteShell:  true,
		WireGuard:    goos == "linux",
	}
}

func CurrentCapabilities() Capabilities {
	return DetectCapabilities(runtime.GOOS, runtime.GOARCH)
}

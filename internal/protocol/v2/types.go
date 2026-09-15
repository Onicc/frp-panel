package v2

import (
	"errors"
	"runtime"
	"time"
)

const Version = "2"

type Hello struct {
	Protocol     string       `json:"protocol"`
	NodeID       string       `json:"nodeId"`
	AgentVersion string       `json:"agentVersion"`
	Capabilities Capabilities `json:"capabilities"`
}

type Capabilities struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	FRPC         bool   `json:"frpc"`
	FRPS         bool   `json:"frps"`
	Functions    bool   `json:"functions"`
	RemoteShell  bool   `json:"remoteShell"`
	WireGuard    bool   `json:"wireGuard"`
}

type DesiredState struct {
	Revision  uint64     `json:"revision"`
	IssuedAt  time.Time  `json:"issuedAt"`
	Resources []Resource `json:"resources"`
}

type Resource struct {
	Kind       string         `json:"kind"`
	ID         string         `json:"id"`
	Generation uint64         `json:"generation"`
	Spec       map[string]any `json:"spec"`
}

type ApplyResult struct {
	Revision uint64          `json:"revision"`
	Applied  bool            `json:"applied"`
	Errors   []ResourceError `json:"errors,omitempty"`
}

type ResourceError struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Heartbeat struct {
	Revision uint64    `json:"revision"`
	SentAt   time.Time `json:"sentAt"`
	Healthy  bool      `json:"healthy"`
}

func (h Hello) Validate() error {
	if h.Protocol != Version {
		return errors.New("unsupported agent protocol")
	}
	if h.NodeID == "" || h.AgentVersion == "" {
		return errors.New("nodeId and agentVersion are required")
	}
	return nil
}

func CurrentCapabilities() Capabilities {
	return Capabilities{
		OS: runtime.GOOS, Architecture: runtime.GOARCH,
		FRPC: true, FRPS: false, Functions: runtime.GOOS != "windows",
		RemoteShell: true, WireGuard: runtime.GOOS == "linux",
	}
}

func (desired DesiredState) ShouldApply(currentRevision uint64) bool {
	return desired.Revision > currentRevision
}

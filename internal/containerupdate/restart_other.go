//go:build !linux

package containerupdate

// Container-managed updates are available only for Linux Docker deployments.
func RestartAfterResponse() {}

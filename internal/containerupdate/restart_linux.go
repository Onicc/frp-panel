//go:build linux

package containerupdate

import (
	"os"
	"syscall"
	"time"
)

func RestartAfterResponse() {
	go func() {
		time.Sleep(750 * time.Millisecond)
		_ = syscall.Kill(os.Getpid(), syscall.SIGTERM)
	}()
}

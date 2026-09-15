//go:build !windows

package agentservice

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// RunManaged translates terminal and service-manager stop signals into context
// cancellation for the Agent runtime.
func RunManaged(run func(context.Context) error) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	return run(ctx)
}

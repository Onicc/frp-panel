//go:build windows

package agentservice

import (
	"context"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sys/windows/svc"
)

// RunManaged registers the Windows Service Control Dispatcher when launched by
// SCM and retains ordinary Ctrl+C behavior for an interactive console.
func RunManaged(run func(context.Context) error) error {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return err
	}
	if isService {
		return svc.Run(ServiceName, &windowsService{run: run})
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	return run(ctx)
}

type windowsService struct {
	run func(context.Context) error
}

func (service *windowsService) Execute(_ []string, requests <-chan svc.ChangeRequest, statuses chan<- svc.Status) (bool, uint32) {
	statuses <- svc.Status{State: svc.StartPending}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- service.run(ctx) }()
	current := svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	statuses <- current

	for {
		select {
		case err := <-done:
			statuses <- svc.Status{State: svc.StopPending}
			if err != nil {
				return true, 1
			}
			return false, 0
		case change := <-requests:
			switch change.Cmd {
			case svc.Interrogate:
				statuses <- current
			case svc.Stop, svc.Shutdown:
				statuses <- svc.Status{State: svc.StopPending}
				cancel()
				select {
				case err := <-done:
					if err != nil {
						return true, 1
					}
					return false, 0
				case <-time.After(30 * time.Second):
					return true, 2
				}
			}
		}
	}
}

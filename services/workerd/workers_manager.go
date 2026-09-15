package workerd

import (
	"fmt"

	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils"
	"github.com/Onicc/frp-panel/utils/logger"
)

type workersManager struct {
	workers *utils.SyncMap[string, app.WorkerController]
}

func NewWorkersManager() *workersManager {
	return &workersManager{
		workers: &utils.SyncMap[string, app.WorkerController]{},
	}
}

func (m *workersManager) GetWorker(ctx *app.Context, id string) (app.WorkerController, bool) {
	return m.workers.Load(id)
}

func (m *workersManager) RunWorker(ctx *app.Context, id string, worker app.WorkerController) error {
	if !ctx.GetApp().GetConfig().Client.Features.EnableFunctions {
		logger.Logger(ctx).Errorf("function features are not enabled")
		return fmt.Errorf("function features are not enabled")
	}

	worker.RunWorker(ctx)

	m.workers.Store(id, worker)
	return nil
}

func (m *workersManager) StopWorker(ctx *app.Context, id string) error {
	worker, ok := m.workers.Load(id)
	if !ok {
		return fmt.Errorf("cannot find worker, id: %s", id)
	}
	worker.StopWorker(ctx)
	m.workers.Delete(id)
	return nil
}

func (m *workersManager) StopAllWorkers(ctx *app.Context) {
	m.workers.Range(func(k string, v app.WorkerController) bool {
		v.StopWorker(ctx)
		return true
	})

	tmpM := m.workers.ToMap()
	for k := range tmpM {
		m.workers.Delete(k)
	}
}

func (m *workersManager) GetWorkerStatus(ctx *app.Context, id string) (defs.WorkerStatus, error) {
	ok, err := utils.ProcessExistsBySelf(id)
	if err != nil {
		return defs.WorkerStatus_Unknown, err
	}
	if ok {
		return defs.WorkerStatus_Running, nil
	}
	return defs.WorkerStatus_Inactive, nil
}

func (m *workersManager) InstallWorkerd(ctx *app.Context, url string, installDir string) (string, error) {
	return "", fmt.Errorf("automatic workerd installation is disabled; install a verified workerd package and set CLIENT_WORKER_WORKERD_BINARY_PATH")
}

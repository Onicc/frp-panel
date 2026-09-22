package client

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Onicc/frp-panel/services/app"
	"github.com/Onicc/frp-panel/utils"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/client/proxy"
	"github.com/fatedier/frp/pkg/config/source"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/fatedier/frp/pkg/config/v1/validation"
	"github.com/fatedier/frp/pkg/policy/featuregate"
	"github.com/fatedier/frp/pkg/policy/security"
	"github.com/samber/lo"
)

type clientImpl struct {
	cli         *client.Service
	Common      *v1.ClientCommonConfig
	ProxyCfgs   map[string]v1.ProxyConfigurer
	VisitorCfgs map[string]v1.VisitorConfigurer
	mu          sync.RWMutex
	done        chan struct{}
	started     bool
	stopped     bool
	running     bool
}

func NewClientHandler(commonCfg *v1.ClientCommonConfig,
	proxyCfgs []v1.ProxyConfigurer,
	visitorCfgs []v1.VisitorConfigurer) app.ClientHandler {
	ctx := context.Background()

	if len(commonCfg.FeatureGates) > 0 {
		if err := featuregate.SetFromMap(commonCfg.FeatureGates); err != nil {
			logger.Logger(ctx).WithError(err).Errorf("there's a feature gate settings, but set failed: %+v, skip", commonCfg.FeatureGates)
		}
	}

	unsafeFeatures := security.NewUnsafeFeatures(nil)
	warning, err := validation.ValidateAllClientConfig(commonCfg, proxyCfgs, visitorCfgs, unsafeFeatures)
	if warning != nil {
		logger.Logger(ctx).WithError(err).Warnf("validate client config warning: %+v", warning)
	}
	if err != nil {
		logger.Logger(ctx).Panic(err)
	}

	configSource := source.NewConfigSource()
	if err := configSource.ReplaceAll(proxyCfgs, visitorCfgs); err != nil {
		logger.Logger(ctx).Panic(err)
	}
	cli, err := client.NewService(client.ServiceOptions{
		Common:                 commonCfg,
		ConfigSourceAggregator: source.NewAggregator(configSource),
		UnsafeFeatures:         unsafeFeatures,
	})
	if err != nil {
		logger.Logger(ctx).Panic(err)
	}

	return &clientImpl{
		cli:         cli,
		Common:      commonCfg,
		ProxyCfgs:   lo.SliceToMap(proxyCfgs, utils.TransformProxyConfigurerToMap),
		VisitorCfgs: lo.SliceToMap(visitorCfgs, utils.TransformVisitorConfigurerToMap),
		done:        make(chan struct{}),
	}
}

func (c *clientImpl) Run() {
	c.mu.Lock()
	if c.started || c.stopped {
		c.mu.Unlock()
		logger.Logger(context.Background()).Warn("client is running, skip run")
		return
	}
	c.started = true
	c.running = true
	c.mu.Unlock()

	shouldGracefulClose := c.Common.Transport.Protocol == "kcp" || c.Common.Transport.Protocol == "quic"
	if shouldGracefulClose {
		go handleTermSignal(c.cli)
	}
	defer func() {
		c.mu.Lock()
		c.running = false
		close(c.done)
		c.mu.Unlock()
	}()
	ctx := context.Background()
	logger.Logger(ctx).Infof("start to run client")
	if err := c.cli.Run(ctx); err != nil {
		logger.Logger(ctx).Errorf("run client error: %v", err)
	}
}

func (c *clientImpl) Stop() {
	c.mu.Lock()
	if c.stopped {
		started := c.started
		c.mu.Unlock()
		if started {
			<-c.done
		}
		return
	}
	c.stopped = true
	started := c.started
	c.mu.Unlock()
	if started {
		c.cli.Close()
		<-c.done
	}
}

func (c *clientImpl) Update(proxyCfgs []v1.ProxyConfigurer, visitorCfgs []v1.VisitorConfigurer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ProxyCfgs = lo.SliceToMap(proxyCfgs, utils.TransformProxyConfigurerToMap)
	c.VisitorCfgs = lo.SliceToMap(visitorCfgs, utils.TransformVisitorConfigurerToMap)
	c.cli.UpdateAllConfigurer(proxyCfgs, visitorCfgs)
}

func (c *clientImpl) AddProxy(proxyCfg v1.ProxyConfigurer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ProxyCfgs[proxyCfg.GetBaseConfig().Name] = proxyCfg
	c.cli.UpdateAllConfigurer(lo.Values(c.ProxyCfgs), lo.Values(c.VisitorCfgs))
}

func (c *clientImpl) AddVisitor(visitorCfg v1.VisitorConfigurer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.VisitorCfgs[visitorCfg.GetBaseConfig().Name] = visitorCfg
	c.cli.UpdateAllConfigurer(lo.Values(c.ProxyCfgs), lo.Values(c.VisitorCfgs))
}

func (c *clientImpl) RemoveProxy(proxyCfg v1.ProxyConfigurer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	old := c.ProxyCfgs
	delete(old, proxyCfg.GetBaseConfig().Name)

	c.ProxyCfgs = old
	c.cli.UpdateAllConfigurer(lo.Values(c.ProxyCfgs), lo.Values(c.VisitorCfgs))
}

func (c *clientImpl) RemoveVisitor(visitorCfg v1.VisitorConfigurer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	old := c.VisitorCfgs
	delete(old, visitorCfg.GetBaseConfig().Name)

	c.VisitorCfgs = old
	c.cli.UpdateAllConfigurer(lo.Values(c.ProxyCfgs), lo.Values(c.VisitorCfgs))
}

func (c *clientImpl) GetProxyStatus(name string) (*proxy.WorkingStatus, bool) {
	return c.cli.StatusExporter().GetProxyStatus(name)
}

func (c *clientImpl) GetCommonCfg() *v1.ClientCommonConfig {
	return c.Common
}

func (c *clientImpl) GetProxyCfgs() map[string]v1.ProxyConfigurer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return lo.Assign(map[string]v1.ProxyConfigurer{}, c.ProxyCfgs)
}

func (c *clientImpl) GetVisitorCfgs() map[string]v1.VisitorConfigurer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return lo.Assign(map[string]v1.VisitorConfigurer{}, c.VisitorCfgs)
}

func (c *clientImpl) Running() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

func (c *clientImpl) Wait() {
	c.mu.RLock()
	started := c.started
	c.mu.RUnlock()
	if started {
		<-c.done
	}
}

func handleTermSignal(svr *client.Service) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	svr.GracefulClose(500 * time.Millisecond)
}

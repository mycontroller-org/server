package helper

import (
	"context"

	"github.com/mycontroller-org/server/v2/pkg/configuration"
	"github.com/mycontroller-org/server/v2/pkg/encryption"
	gatewaySVC "github.com/mycontroller-org/server/v2/pkg/service/gateway"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

type Gateway struct {
	ctx    context.Context
	config *config.Config
	logger *zap.Logger
	enc    *encryption.Encryption
	bus    busTY.Plugin

	// services
	coreSchedulerSVC schedulerTY.CoreScheduler // core scheduler, used to execute all the cron jobs
	gatewaySVC       serviceTY.Service
}

func (g *Gateway) Start(ctx context.Context, configFilePath string) error {
	g.ctx = ctx

	// load config
	cfg, err := configuration.Get(configFilePath)
	if err != nil {
		return err
	}

	// load logger
	ctx, logger := loadLogger(ctx, cfg.Logger, "gateway")

	// set environment values
	err = setEnvironmentVariables(cfg)
	if err != nil {
		logger.Error("error on setting environment variables", zap.Error(err))
		return err
	}

	// load encryption helper
	ctx, enc := loadEncryptionHelper(ctx, logger, cfg.Secret)

	// get core scheduler and inject into context
	ctx, coreScheduler := loadCoreScheduler(ctx)
	err = coreScheduler.Start()
	if err != nil {
		logger.Error("error on starting core scheduler", zap.Error(err))
		return err
	}

	// load bus plugin
	ctx, bus, err := loadBus(ctx, cfg.Bus)
	if err != nil {
		logger.Error("error on getting bus", zap.Error(err))
		return err
	}

	// add into struct
	g.ctx = ctx
	g.config = cfg
	g.logger = logger
	g.enc = enc
	g.coreSchedulerSVC = coreScheduler
	g.bus = bus

	// load services
	//-------------------------------------------

	// gateway service
	gateway, err := gatewaySVC.New(ctx, &cfg.Gateway)
	if err != nil {
		logger.Error("error on getting gateway service", zap.Error(err))
		return err
	}

	err = gateway.Start()
	if err != nil {
		g.logger.Error("error on starting a service", zap.String("name", gateway.Name()), zap.Error(err))
		return err
	}

	// add services into struct
	g.gatewaySVC = gateway

	// call shutdown hook
	shutdownHook := NewShutdownHook(g.logger, g.stop, g.bus, false)

	shutdownHook.Start()

	return nil
}

func (g *Gateway) stop() {
	// stop services
	if err := g.gatewaySVC.Close(); err != nil {
		g.logger.Error("error on closing a service", zap.String("name", g.gatewaySVC.Name()), zap.Error(err))
	}

	// unload support apis and plugins
	if err := g.bus.Close(); err != nil {
		g.logger.Error("error on closing bus service", zap.Error(err))
	}
	if err := g.coreSchedulerSVC.Close(); err != nil {
		g.logger.Error("error on closing core scheduler", zap.Error(err))
	}
}

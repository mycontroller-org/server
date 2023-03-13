package helper

import (
	"context"

	"github.com/mycontroller-org/server/v2/pkg/configuration"
	"github.com/mycontroller-org/server/v2/pkg/encryption"
	handlerSVC "github.com/mycontroller-org/server/v2/pkg/service/handler"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

type Handler struct {
	ctx    context.Context
	config *config.Config
	logger *zap.Logger
	enc    *encryption.Encryption
	bus    busTY.Plugin

	// services
	coreSchedulerSVC schedulerTY.CoreScheduler // core scheduler, used to execute all the cron jobs
	handlerSVC       serviceTY.Service
}

func (h Handler) Start(ctx context.Context, configFilePath string) error {
	h.ctx = ctx

	// load config
	cfg, err := configuration.Get(configFilePath)
	if err != nil {
		return err
	}

	// load logger
	ctx, logger := loadLogger(ctx, cfg.Logger, "handler")

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
	h.ctx = ctx
	h.config = cfg
	h.logger = logger
	h.enc = enc
	h.coreSchedulerSVC = coreScheduler
	h.bus = bus

	// load services
	//-------------------------------------------

	// handler service
	handler, err := handlerSVC.New(ctx, &cfg.Gateway)
	if err != nil {
		logger.Error("error on getting gateway service", zap.Error(err))
		return err
	}

	err = handler.Start()
	if err != nil {
		h.logger.Error("error on starting a service", zap.String("name", handler.Name()), zap.Error(err))
		return err
	}

	// add services into struct
	h.handlerSVC = handler

	// call shutdown hook
	shutdownHook := NewShutdownHook(h.logger, h.stop, h.bus, false)

	shutdownHook.Start()

	return nil
}

func (h Handler) stop() {
	// stop services
	if err := h.handlerSVC.Close(); err != nil {
		h.logger.Error("error on closing a service", zap.String("name", h.handlerSVC.Name()), zap.Error(err))
	}

	// unload support apis and plugins
	if err := h.bus.Close(); err != nil {
		h.logger.Error("error on closing bus service", zap.Error(err))
	}
	if err := h.coreSchedulerSVC.Close(); err != nil {
		h.logger.Error("error on closing core scheduler", zap.Error(err))
	}
}

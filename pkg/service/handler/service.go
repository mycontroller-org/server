package handler

import (
	"fmt"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	handlerPlugin "github.com/mycontroller-org/server/v2/plugin/handler"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// startHandler notify handlers
func (svc *HandlerService) startHandler(cfg *handlerTY.Config) error {
	if svc.store.Get(cfg.ID) != nil {
		return fmt.Errorf("a service is in running state. id:%s", cfg.ID)
	}
	if !cfg.Enabled { // this handler is not enabled
		return nil
	}
	svc.logger.Debug("starting a handler", zap.Any("id", cfg.ID))
	state := types.State{Since: time.Now()}

	handler, err := svc.loadHandler(cfg)
	if err != nil {
		return err
	}
	err = handler.Start()
	if err != nil {
		svc.logger.Error("unable to start a handler service", zap.Any("id", cfg.ID), zap.Error(err))
		state.Message = err.Error()
		state.Status = types.StatusDown
	} else {
		state.Message = "started successfully"
		state.Status = types.StatusUp
		svc.store.Add(cfg.ID, handler)
	}

	busUtils.SetHandlerState(svc.logger, svc.bus, cfg.ID, state)
	return nil
}

// stopHandler a handler
func (svc *HandlerService) stopHandler(id string) error {
	svc.logger.Debug("stopping a handler", zap.Any("id", id))
	handler := svc.store.Get(id)
	if handler != nil {
		err := handler.Close()
		state := types.State{
			Status:  types.StatusDown,
			Since:   time.Now(),
			Message: "stopped by request",
		}
		if err != nil {
			svc.logger.Error("failed to stop handler service", zap.String("id", id), zap.Error(err))
			state.Message = err.Error()
		}
		busUtils.SetHandlerState(svc.logger, svc.bus, id, state)
		svc.store.Remove(id)
	}
	return nil
}

// reloadHandler a handler
func (svc *HandlerService) reloadHandler(gwCfg *handlerTY.Config) error {
	err := svc.stopHandler(gwCfg.ID)
	if err != nil {
		return err
	}
	utils.SmartSleep(1 * time.Second)
	return svc.startHandler(gwCfg)
}

// unloadAll stops all handlers
func (svc *HandlerService) unloadAll() {
	ids := svc.store.ListIDs()
	for _, id := range ids {
		err := svc.stopHandler(id)
		if err != nil {
			svc.logger.Error("error on stopping a handler", zap.String("id", id))
		}
	}
}

func (svc *HandlerService) loadHandler(cfg *handlerTY.Config) (handlerTY.Plugin, error) {
	// decrypt secrets, tokens
	err := svc.enc.DecryptSecrets(cfg)
	if err != nil {
		return nil, err
	}

	handler, err := handlerPlugin.Create(svc.ctx, cfg.Type, cfg)
	if err != nil {
		return nil, err
	}
	return handler, nil
}

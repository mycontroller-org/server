package service

import (
	"fmt"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	vaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/types"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	vaPlugin "github.com/mycontroller-org/server/v2/plugin/virtual_assistant"
	"go.uber.org/zap"
)

// Start a virtual assistant
func (svc *VirtualAssistantService) startAssistant(cfg *vaTY.Config) error {
	start := time.Now()

	// decrypt the secrets, tokens
	err := svc.enc.DecryptSecrets(cfg)
	if err != nil {
		return err
	}

	if svc.store.Get(cfg.ID) != nil {
		svc.logger.Info("no action needed. virtual assistant service is in running state.", zap.String("id", cfg.ID))
		return nil
	}
	if !cfg.Enabled { // this assistant is not enabled
		return nil
	}
	svc.logger.Info("starting a virtual assistant", zap.Any("id", cfg.ID))
	state := types.State{Since: time.Now()}

	service, err := vaPlugin.Create(svc.ctx, cfg.ProviderType, cfg)
	if err != nil {
		return err
	}
	err = service.Start()
	if err != nil {
		svc.logger.Error("failed to start a virtual assistant", zap.String("id", cfg.ID), zap.String("timeTaken", time.Since(start).String()), zap.Error(err))
		state.Message = err.Error()
		state.Status = types.StatusDown
	} else {
		svc.logger.Info("started a virtual assistant", zap.String("id", cfg.ID), zap.String("timeTaken", time.Since(start).String()))
		state.Message = "Started successfully"
		state.Status = types.StatusUp
		svc.store.Add(service)
	}

	busUtils.SetVirtualAssistantState(svc.logger, svc.bus, cfg.ID, state)
	return nil
}

// stop a assistant
func (svc *VirtualAssistantService) stopAssistant(id string) error {
	start := time.Now()
	svc.logger.Info("stopping a virtual assistant", zap.Any("id", id))
	service := svc.store.Get(id)
	if service != nil {
		err := service.Stop()
		state := types.State{
			Status:  types.StatusDown,
			Since:   time.Now(),
			Message: "Stopped by request",
		}
		if err != nil {
			svc.logger.Error("failed to stop a virtual assistant", zap.String("id", id), zap.String("timeTaken", time.Since(start).String()), zap.Error(err))
			state.Message = fmt.Sprintf("Failed to stop: %s", err.Error())
			busUtils.SetVirtualAssistantState(svc.logger, svc.bus, id, state)
		} else {
			svc.logger.Info("stopped a virtual assistant", zap.String("id", id), zap.String("timeTaken", time.Since(start).String()))
			busUtils.SetVirtualAssistantState(svc.logger, svc.bus, id, state)
			svc.store.Remove(id)
		}
	}
	return nil
}

// unloadAll stops all assistants
func (svc *VirtualAssistantService) unloadAll() {
	ids := svc.store.ListIDs()
	for _, id := range ids {
		err := svc.stopAssistant(id)
		if err != nil {
			svc.logger.Error("error on stopping a virtual assistant", zap.String("id", id))
		}
	}
}

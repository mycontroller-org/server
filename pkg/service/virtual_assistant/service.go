package service

import (
	"fmt"
	"time"

	commonStore "github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	vaTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_assistant"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	vaPlugin "github.com/mycontroller-org/server/v2/plugin/virtual_assistant"
	"go.uber.org/zap"
)

// Start a virtual assistant
func StartAssistant(cfg *vaTY.Config) error {
	start := time.Now()

	// decrypt the secrets, tokens
	err := cloneUtil.UpdateSecrets(cfg, commonStore.CFG.Secret, "", false, cloneUtil.DefaultSpecialKeys)
	if err != nil {
		return err
	}

	if vaService.Get(cfg.ID) != nil {
		zap.L().Info("no action needed. virtual assistant service is in running state.", zap.String("id", cfg.ID))
		return nil
	}
	if !cfg.Enabled { // this assistant is not enabled
		return nil
	}
	zap.L().Info("starting a virtual assistant", zap.Any("id", cfg.ID))
	state := types.State{Since: time.Now()}

	service, err := vaPlugin.Create(cfg.ProviderType, cfg)
	if err != nil {
		return err
	}
	err = service.Start()
	if err != nil {
		zap.L().Error("failed to start a virtual assistant", zap.String("id", cfg.ID), zap.String("timeTaken", time.Since(start).String()), zap.Error(err))
		state.Message = err.Error()
		state.Status = types.StatusDown
	} else {
		zap.L().Info("started a virtual assistant", zap.String("id", cfg.ID), zap.String("timeTaken", time.Since(start).String()))
		state.Message = "Started successfully"
		state.Status = types.StatusUp
		vaService.Add(service)
	}

	busUtils.SetVirtualAssistantState(cfg.ID, state)
	return nil
}

// stop a assistant
func StopAssistant(id string) error {
	start := time.Now()
	zap.L().Info("stopping a virtual assistant", zap.Any("id", id))
	service := vaService.Get(id)
	if service != nil {
		err := service.Stop()
		state := types.State{
			Status:  types.StatusDown,
			Since:   time.Now(),
			Message: "Stopped by request",
		}
		if err != nil {
			zap.L().Error("failed to stop a virtual assistant", zap.String("id", id), zap.String("timeTaken", time.Since(start).String()), zap.Error(err))
			state.Message = fmt.Sprintf("Failed to stop: %s", err.Error())
			busUtils.SetVirtualAssistantState(id, state)
		} else {
			zap.L().Info("stopped a virtual assistant", zap.String("id", id), zap.String("timeTaken", time.Since(start).String()))
			busUtils.SetVirtualAssistantState(id, state)
			vaService.Remove(id)
		}
	}
	return nil
}

// reload a assistant
func ReloadAssistant(gwCfg *vaTY.Config) error {
	err := StopAssistant(gwCfg.ID)
	if err != nil {
		return err
	}
	utils.SmartSleep(1 * time.Second)
	return StartAssistant(gwCfg)
}

// UnloadAll stops all assistants
func UnloadAll() {
	ids := vaService.ListIDs()
	for _, id := range ids {
		err := StopAssistant(id)
		if err != nil {
			zap.L().Error("error on stopping a virtual assistant", zap.String("id", id))
		}
	}
}

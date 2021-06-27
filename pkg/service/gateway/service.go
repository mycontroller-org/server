package service

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/model"
	gwML "github.com/mycontroller-org/server/v2/pkg/model/gateway"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	gwProvider "github.com/mycontroller-org/server/v2/plugin/gateway/provider"
	"go.uber.org/zap"
)

// StartGW gateway
func StartGW(gatewayCfg *gwML.Config) error {
	start := time.Now()

	// descrypt the secrets, tokens
	err := cloneUtil.UpdateSecrets(gatewayCfg, false)
	if err != nil {
		return err
	}

	if gwService.Get(gatewayCfg.ID) != nil {
		zap.L().Info("no action needed. gateway service is in running state.", zap.String("gatewayId", gatewayCfg.ID))
		return nil
	}
	if !gatewayCfg.Enabled { // this gateway is not enabled
		return nil
	}
	zap.L().Info("starting a gateway", zap.Any("id", gatewayCfg.ID))
	state := model.State{Since: time.Now()}

	service, err := gwProvider.GetService(gatewayCfg)
	if err != nil {
		return err
	}
	err = service.Start()
	if err != nil {
		zap.L().Error("failed to start a gateway", zap.String("id", gatewayCfg.ID), zap.String("timeTaken", time.Since(start).String()), zap.Error(err))
		state.Message = err.Error()
		state.Status = model.StatusDown
	} else {
		zap.L().Info("started a gateway", zap.String("id", gatewayCfg.ID), zap.String("timeTaken", time.Since(start).String()))
		state.Message = "Started successfully"
		state.Status = model.StatusUp
		gwService.Add(service)
	}

	busUtils.SetGatewayState(gatewayCfg.ID, state)
	return nil
}

// StopGW gateway
func StopGW(id string) error {
	start := time.Now()
	zap.L().Info("stopping a gateway", zap.Any("id", id))
	service := gwService.Get(id)
	if service != nil {
		err := service.Stop()
		state := model.State{
			Status:  model.StatusDown,
			Since:   time.Now(),
			Message: "Stopped by request",
		}
		if err != nil {
			zap.L().Error("failed to stop a gateway", zap.String("id", id), zap.String("timeTaken", time.Since(start).String()), zap.Error(err))
			state.Message = fmt.Sprintf("Failed to stop: %s", err.Error())
			busUtils.SetGatewayState(id, state)
		} else {
			zap.L().Info("stopped a gateway", zap.String("id", id), zap.String("timeTaken", time.Since(start).String()))
			busUtils.SetGatewayState(id, state)
			gwService.Remove(id)
		}
	}
	return nil
}

// ReloadGW gateway
func ReloadGW(gwCfg *gwML.Config) error {
	err := StopGW(gwCfg.ID)
	if err != nil {
		return err
	}
	return StartGW(gwCfg)
}

// UnloadAll stops all the gateways
func UnloadAll() {
	ids := gwService.ListIDs()
	for _, id := range ids {
		err := StopGW(id)
		if err != nil {
			zap.L().Error("error on stopping a gateway", zap.String("id", id))
		}
	}
}

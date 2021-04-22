package service

import (
	"fmt"
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	gwpd "github.com/mycontroller-org/backend/v2/plugin/gateway/provider"
	"go.uber.org/zap"
)

// Start gateway
func Start(gatewayCfg *gwml.Config) error {
	start := time.Now()
	if gwService.Get(gatewayCfg.ID) != nil {
		return fmt.Errorf("a service is in running state. gateway:%s", gatewayCfg.ID)
	}
	if !gatewayCfg.Enabled { // this gateway is not enabled
		return nil
	}
	zap.L().Info("starting a gateway", zap.Any("id", gatewayCfg.ID))
	state := ml.State{Since: time.Now()}

	service, err := gwpd.GetService(gatewayCfg)
	if err != nil {
		return err
	}
	err = service.Start()
	if err != nil {
		zap.L().Error("failed to start a gateway", zap.String("id", gatewayCfg.ID), zap.String("timeTaken", time.Since(start).String()), zap.Error(err))
		state.Message = err.Error()
		state.Status = ml.StatusDown
	} else {
		zap.L().Info("started a gateway", zap.String("id", gatewayCfg.ID), zap.String("timeTaken", time.Since(start).String()))
		state.Message = "Started successfully"
		state.Status = ml.StatusUp
		gwService.Add(service)
	}

	busUtils.SetGatewayState(gatewayCfg.ID, state)
	return nil
}

// Stop gateway
func Stop(id string) error {
	start := time.Now()
	zap.L().Info("stopping a gateway", zap.Any("id", id))
	service := gwService.Get(id)
	if service != nil {
		err := service.Stop()
		state := ml.State{
			Status:  ml.StatusDown,
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

// Reload gateway
func Reload(gwCfg *gwml.Config) error {
	err := Stop(gwCfg.ID)
	if err != nil {
		return err
	}
	return Start(gwCfg)
}

// UnloadAll stops all the gateways
func UnloadAll() {
	ids := gwService.ListIDs()
	for _, id := range ids {
		err := Stop(id)
		if err != nil {
			zap.L().Error("error on stopping a gateway", zap.String("id", id))
		}
	}
}

package service

import (
	"fmt"
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	rsUtils "github.com/mycontroller-org/backend/v2/pkg/utils/resource_service"
	gwpd "github.com/mycontroller-org/backend/v2/plugin/gateway/provider"
	"go.uber.org/zap"
)

// Start gateway
func Start(gatewayCfg *gwml.Config) error {
	start := time.Now()
	if gwService.Get(gatewayCfg.ID) != nil {
		return fmt.Errorf("A service is in running state. gateway:%s", gatewayCfg.ID)
	}
	if !gatewayCfg.Enabled { // this gateway is not enabled
		return nil
	}
	zap.L().Info("Starting a gateway", zap.Any("id", gatewayCfg.ID))
	state := ml.State{Since: time.Now()}

	service, err := gwpd.GetService(gatewayCfg)
	if err != nil {
		return err
	}
	err = service.Start()
	if err != nil {
		zap.L().Error("Failed to start a gateway", zap.String("id", gatewayCfg.ID), zap.String("timeTaken", time.Since(start).String()), zap.Error(err))
		state.Message = err.Error()
		state.Status = ml.StateDown
	} else {
		zap.L().Info("Started a gateway", zap.String("id", gatewayCfg.ID), zap.String("timeTaken", time.Since(start).String()))
		state.Message = "Started successfully"
		state.Status = ml.StateUp
		gwService.Add(service)
	}

	rsUtils.SetGatewayState(gatewayCfg.ID, state)
	return nil
}

// Stop gateway
func Stop(id string) error {
	start := time.Now()
	zap.L().Info("Stopping a gateway", zap.Any("id", id))
	service := gwService.Get(id)
	if service != nil {
		err := service.Stop()
		state := ml.State{
			Status:  ml.StateDown,
			Since:   time.Now(),
			Message: "Stopped by request",
		}
		if err != nil {
			zap.L().Error("Failed to stop a gateway", zap.String("id", id), zap.String("timeTaken", time.Since(start).String()), zap.Error(err))
			state.Message = fmt.Sprintf("Failed to stop: %s", err.Error())
			rsUtils.SetGatewayState(id, state)
		} else {
			zap.L().Info("Stopped a gateway", zap.String("id", id), zap.String("timeTaken", time.Since(start).String()))
			rsUtils.SetGatewayState(id, state)
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

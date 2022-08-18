package service

import (
	"fmt"
	"time"

	commonStore "github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	gwProvider "github.com/mycontroller-org/server/v2/plugin/gateway/provider"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

// StartGW gateway
func StartGW(gatewayCfg *gwTY.Config) error {
	start := time.Now()

	// decrypt the secrets, tokens
	err := cloneUtil.UpdateSecrets(gatewayCfg, commonStore.CFG.Secret, "", false, cloneUtil.DefaultSpecialKeys)
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
	state := types.State{Since: time.Now()}

	service, err := gwProvider.GetService(gatewayCfg)
	if err != nil {
		return err
	}
	err = service.Start()
	if err != nil {
		zap.L().Error("failed to start a gateway", zap.String("id", gatewayCfg.ID), zap.String("timeTaken", time.Since(start).String()), zap.Error(err))
		state.Message = err.Error()
		state.Status = types.StatusDown
	} else {
		zap.L().Info("started a gateway", zap.String("id", gatewayCfg.ID), zap.String("timeTaken", time.Since(start).String()))
		state.Message = "Started successfully"
		state.Status = types.StatusUp
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
		state := types.State{
			Status:  types.StatusDown,
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
func ReloadGW(gwCfg *gwTY.Config) error {
	err := StopGW(gwCfg.ID)
	if err != nil {
		return err
	}
	utils.SmartSleep(1 * time.Second)
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

// returns sleeping queue messages from the given gateway ID
func getGatewaySleepingQueue(gatewayID string) *map[string][]msgTY.Message {
	service := gwService.Get(gatewayID)
	if service != nil {
		messages := service.GetGatewaySleepingQueue()
		return &messages
	}
	return nil
}

// returns sleeping queue messages from the given gateway ID and node ID
func getNodeSleepingQueue(gatewayID, nodeID string) *[]msgTY.Message {
	service := gwService.Get(gatewayID)
	if service != nil {
		messages := service.GetNodeSleepingQueue(nodeID)
		return &messages
	}
	return nil
}

// clears sleeping queue of a gateway
func clearGatewaySleepingQueue(gatewayID string) {
	service := gwService.Get(gatewayID)
	if service != nil {
		service.ClearGatewaySleepingQueue()
	}
}

// clears sleeping queue of a node
func clearNodeSleepingQueue(gatewayID, nodeID string) {
	service := gwService.Get(gatewayID)
	if service != nil {
		service.ClearNodeSleepingQueue(nodeID)
	}
}

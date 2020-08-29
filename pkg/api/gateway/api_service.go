package gateway

import (
	"context"
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	gwpd "github.com/mycontroller-org/backend/v2/plugin/gateway_provider"
	"github.com/mycontroller-org/backend/v2/plugin/gateway_provider/mysensors"
	"go.uber.org/zap"
)

// Start gateway
func Start(gwCfg *gwml.Config) error {
	var provider Provider
	switch gwCfg.Provider.Type {
	case gwpd.ProviderMySensors:
		provider = &mysensors.Provider{GWConfig: gwCfg}
	default:
		zap.L().Info("Unknown provider", zap.Any("gateway", gwCfg))
		return nil
	}
	s := &Service{
		Config:   gwCfg,
		Provider: provider,
		ctx:      context.TODO(),
	}

	err := s.Start()
	if err != nil {
		zap.L().Info("Unable to start the gateway", zap.Any("gateway", gwCfg))
	} else {
		AddGatewayService(s)
	}

	return nil
}

// Stop gateway
func Stop(gwCfg *gwml.Config) error {
	gs := GetGatewayService(gwCfg)
	if gs != nil {
		err := gs.Stop()
		if err != nil {
			zap.L().Error("Failed to stop media service", zap.Error(err))
		}
	}
	return nil
}

// Reload gateway
func Reload(gwCfg *gwml.Config) error {
	err := Stop(gwCfg)
	if err != nil {
		return err
	}
	if gwCfg.Enabled {
		err = Start(gwCfg)
	}
	return err
}

// LoadGateways makes gateways alive
func LoadGateways() {
	gws, err := ListGateways(nil, pml.Pagination{})
	if err != nil {
		zap.L().Error("Error getting list of gateways", zap.Error(err))
	}
	for _, gw := range gws {
		state := ml.State{
			Since:   time.Now(),
			Status:  ml.StateUp,
			Message: "Started successfully",
		}
		err = Start(&gw)
		if err != nil {
			state.Message = err.Error()
			state.Status = ml.StateDown
			zap.L().Error("Error loading a gateway", zap.Error(err), zap.Any("gateway", gw))
		}
		err = SetState(&gw, state)
		if err != nil {
			zap.L().Error("Failed to update gateway status")
		}
	}
}

// UnloadGateways makes stop gateways
func UnloadGateways() {
	gws, err := ListGateways(nil, pml.Pagination{})
	if err != nil {
		zap.L().Error("Error getting list of gateways", zap.Error(err))
	}
	for _, gw := range gws {
		err = Stop(&gw)
		if err != nil {
			zap.L().Error("Error unloading a gateway", zap.Error(err), zap.Any("gateway", gw))
		}
	}
}

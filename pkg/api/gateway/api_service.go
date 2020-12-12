package gateway

import (
	"context"
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	gwpd "github.com/mycontroller-org/backend/v2/plugin/gw_provider"
	"github.com/mycontroller-org/backend/v2/plugin/gw_provider/mysensors"
	"github.com/mycontroller-org/backend/v2/plugin/gw_provider/tasmota"
	"go.uber.org/zap"
)

// Start gateway
func Start(gwCfg *gwml.Config) error {
	zap.L().Debug("Loading gateway", zap.Any("gateway", gwCfg))
	state := ml.State{Since: time.Now()}

	var provider Provider
	switch gwCfg.Provider.Type {

	case gwpd.ProviderMySensors:
		provider = &mysensors.Provider{GWConfig: gwCfg}

	case gwpd.ProviderTasmota:
		provider = &tasmota.Provider{GWConfig: gwCfg}

	default:
		zap.L().Info("Unknown provider", zap.Any("gateway", gwCfg))
		return nil
	}
	service := &Service{
		Config:   gwCfg,
		Provider: provider,
		ctx:      context.TODO(),
	}

	err := service.Start()
	if err != nil {
		zap.L().Error("Unable to start the gateway", zap.Any("gateway", gwCfg), zap.Error(err))
		state.Message = err.Error()
		state.Status = ml.StateDown
	} else {
		state.Message = "Started successfully"
		state.Status = ml.StateUp
		AddGatewayService(service)
	}

	if err := SetState(gwCfg, state); err != nil {
		zap.L().Error("Failed to update gateway state", zap.Error(err))
	}
	return err
}

// Stop gateway
func Stop(gwCfg *gwml.Config) error {
	gatewayService := GetGatewayService(gwCfg)
	if gatewayService != nil {
		err := gatewayService.Stop()
		state := ml.State{
			Status: ml.StateDown,
			Since:  time.Now(),
		}
		if err != nil {
			zap.L().Error("Failed to stop media service", zap.Error(err))
			state.Message = err.Error()
		}
		err = SetState(gwCfg, state)
		if err != nil {
			zap.L().Error("Failed to update gateway state", zap.Error(err))
		}
	}
	return nil
}

// LoadGateways makes gateways alive
func LoadGateways() {
	gwsResult, err := List(nil, nil)
	if err != nil {
		zap.L().Error("Error getting list of gateways", zap.Error(err))
		return
	}
	gateways := *gwsResult.Data.(*[]gwml.Config)
	for index := 0; index < len(gateways); index++ {
		gateway := gateways[index]
		if gateway.Enabled {
			Start(&gateway)
		}
	}
}

// UnloadGateways makes stop gateways
func UnloadGateways() {
	gwsResult, err := List(nil, nil)
	if err != nil {
		zap.L().Error("Error getting list of gateways", zap.Error(err))
	}
	gateways := *gwsResult.Data.(*[]gwml.Config)
	for index := 0; index < len(gateways); index++ {
		gateway := gateways[index]
		if gateway.Enabled {
			err = Stop(&gateway)
			if err != nil {
				zap.L().Error("Error unloading a gateway", zap.Error(err), zap.Any("gateway", gateway))
			}
		}
	}
}

// Enable gateway
func Enable(ID string) error {
	gwCfg, err := Get([]pml.Filter{{Key: ml.KeyID, Value: ID}})
	if err != nil {
		return err
	}
	if !gwCfg.Enabled {
		gwCfg.Enabled = true
		err = Save(&gwCfg)
		if err != nil {
			return err
		}
		return Start(&gwCfg)
	}
	return nil
}

// Disable gateway
func Disable(ID string) error {
	gwCfg, err := Get([]pml.Filter{{Key: ml.KeyID, Value: ID}})
	if err != nil {
		return err
	}
	if gwCfg.Enabled {
		gwCfg.Enabled = false
		err = Save(&gwCfg)
		if err != nil {
			return err
		}
		return Stop(&gwCfg)
	}
	return nil
}

// Reload gateway
func Reload(ID string) error {
	gwCfg, err := Get([]pml.Filter{{Key: ml.KeyID, Value: ID}})
	if err != nil {
		return err
	}
	err = Stop(&gwCfg)
	if err != nil {
		return err
	}
	if gwCfg.Enabled {
		err = Start(&gwCfg)
	}
	return err
}

package gateway

import (
	"fmt"
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	gwpd "github.com/mycontroller-org/backend/v2/plugin/gw_provider"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

// Start gateway
func Start(gatewayCfg *gwml.Config) error {
	if gwService.Get(gatewayCfg) != nil {
		return fmt.Errorf("A service is in running state. gateway:%s", gatewayCfg.ID)
	}
	zap.L().Debug("Starting a gateway", zap.Any("name", gatewayCfg.Name))
	state := ml.State{Since: time.Now()}

	service, err := gwpd.GetService(gatewayCfg)
	if err != nil {
		return err
	}
	err = service.Start()
	if err != nil {
		zap.L().Error("Unable to start a gateway service", zap.Any("name", gatewayCfg.Name), zap.Error(err))
		state.Message = err.Error()
		state.Status = ml.StateDown
	} else {
		state.Message = "Started successfully"
		state.Status = ml.StateUp
		gwService.Add(service)
	}

	if err := SetState(gatewayCfg, state); err != nil {
		zap.L().Error("Failed to update gateway state", zap.String("name", gatewayCfg.Name), zap.Error(err))
	}
	return err
}

// Stop gateway
func Stop(gatewayCfg *gwml.Config) error {
	zap.L().Debug("Stopping a gateway", zap.Any("name", gatewayCfg.Name))
	service := gwService.Get(gatewayCfg)
	if service != nil {
		err := service.Stop()
		state := ml.State{
			Status:  ml.StateDown,
			Since:   time.Now(),
			Message: "Stopped by request",
		}
		if err != nil {
			zap.L().Error("Failed to stop gateway service", zap.String("name", gatewayCfg.Name), zap.Error(err))
			state.Message = err.Error()
		}
		err = SetState(gatewayCfg, state)
		if err != nil {
			zap.L().Error("Failed to update gateway state", zap.String("name", gatewayCfg.Name), zap.Error(err))
		}
		gwService.Remove(gatewayCfg)
	}
	return nil
}

// LoadAll makes gateways alive
func LoadAll() {
	gwsResult, err := List(nil, nil)
	if err != nil {
		zap.L().Error("Failed to get list of gateways", zap.Error(err))
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

// UnloadAll makes stop gateways
func UnloadAll() {
	gwsResult, err := List(nil, nil)
	if err != nil {
		zap.L().Error("Failed to get list of gateways", zap.Error(err))
	}
	gateways := *gwsResult.Data.(*[]gwml.Config)
	for index := 0; index < len(gateways); index++ {
		gateway := gateways[index]
		if gateway.Enabled {
			err = Stop(&gateway)
			if err != nil {
				zap.L().Error("Failed to unload a gateway", zap.Any("name", gateway.Name), zap.Error(err))
			}
		}
	}
}

// Enable gateway
func Enable(ID string) error {
	gwCfg, err := Get([]stgml.Filter{{Key: ml.KeyID, Value: ID}})
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
	gatewayCfg, err := Get([]stgml.Filter{{Key: ml.KeyID, Value: ID}})
	if err != nil {
		return err
	}
	if gatewayCfg.Enabled {
		gatewayCfg.Enabled = false
		err = Save(&gatewayCfg)
		if err != nil {
			return err
		}
		return Stop(&gatewayCfg)
	}
	return nil
}

// Reload gateway
func Reload(ID string) error {
	gwCfg, err := Get([]stgml.Filter{{Key: ml.KeyID, Value: ID}})
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

package gateway

import (
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	rsml "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

// Start gateway
func Start(gwCfg *gwml.Config) error {
	return postGatewayCommand(gwCfg, rsml.CommandStart)
}

// Stop gateway
func Stop(gwCfg *gwml.Config) error {
	return postGatewayCommand(gwCfg, rsml.CommandStop)
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
			err = Start(&gateway)
			if err != nil {
				zap.L().Error("Failed to load a gateway", zap.Error(err), zap.String("gateway", gateway.ID))
			}
		}
	}
}

// UnloadAll makes stop all gateways
func UnloadAll() {
	err := postGatewayCommand(nil, rsml.CommandUnloadAll)
	if err != nil {
		zap.L().Error("error on unload gateways command", zap.Error(err))
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
		err = Save(gwCfg)
		if err != nil {
			return err
		}
		return postGatewayCommand(gwCfg, rsml.CommandStart)
	}
	return nil
}

// Disable gateway
func Disable(ID string) error {
	gwCfg, err := Get([]stgml.Filter{{Key: ml.KeyID, Value: ID}})
	if err != nil {
		return err
	}
	if gwCfg.Enabled {
		gwCfg.Enabled = false
		err = Save(gwCfg)
		if err != nil {
			return err
		}
		return postGatewayCommand(gwCfg, rsml.CommandStop)
	}
	return nil
}

// Reload gateway
func Reload(ID string) error {
	gwCfg, err := Get([]stgml.Filter{{Key: ml.KeyID, Value: ID}})
	if err != nil {
		return err
	}
	return postGatewayCommand(gwCfg, rsml.CommandReload)
}

func postGatewayCommand(gwCfg *gwml.Config, command string) error {
	reqEvent := rsml.Event{
		Type:    rsml.TypeGateway,
		Command: command,
	}
	if gwCfg != nil {
		reqEvent.ID = gwCfg.ID
		err := reqEvent.SetData(gwCfg)
		if err != nil {
			return err
		}
	}
	topic := mcbus.FormatTopic(mcbus.TopicServiceGateway)
	return mcbus.Publish(topic, reqEvent)
}

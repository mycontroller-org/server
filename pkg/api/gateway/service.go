package gateway

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	rsml "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	cloneutil "github.com/mycontroller-org/backend/v2/pkg/utils/clone"
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
func Enable(ids []string) error {
	gateways, err := getGatewayEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(gateways); index++ {
		gwCfg := gateways[index]
		if !gwCfg.Enabled {
			gwCfg.Enabled = true
			err = Save(&gwCfg)
			if err != nil {
				return err
			}
			return Start(&gwCfg)
		}
	}
	return nil
}

// Disable gateway
func Disable(ids []string) error {
	gateways, err := getGatewayEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(gateways); index++ {
		gwCfg := gateways[index]
		if gwCfg.Enabled {
			gwCfg.Enabled = false
			err = Save(&gwCfg)
			if err != nil {
				return err
			}
			return Stop(&gwCfg)
		}
	}
	return nil
}

// Reload gateway
func Reload(ids []string) error {
	gateways, err := getGatewayEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(gateways); index++ {
		gateway := gateways[index]
		err = Stop(&gateway)
		if err != nil {
			zap.L().Error("error on stoping a gateway command", zap.Error(err), zap.String("gateway", gateway.ID))
		}
		if gateway.Enabled {
			err = Start(&gateway)
			if err != nil {
				zap.L().Error("error on start a gateway command", zap.Error(err), zap.String("gateway", gateway.ID))
			}
		}
	}
	return nil
}

func postGatewayCommand(gwCfg *gwml.Config, command string) error {
	reqEvent := rsml.ServiceEvent{
		Type:    rsml.TypeGateway,
		Command: command,
	}
	if gwCfg != nil {
		// descrypt the secrets
		err := cloneutil.UpdateSecrets(gwCfg, false)
		if err != nil {
			return err
		}
		reqEvent.ID = gwCfg.ID
		reqEvent.SetData(gwCfg)
	}
	topic := mcbus.FormatTopic(mcbus.TopicServiceGateway)
	return mcbus.Publish(topic, reqEvent)
}

func getGatewayEntries(ids []string) ([]gwml.Config, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: ids}}
	pagination := &stgml.Pagination{Limit: 100}
	gwsResult, err := List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *gwsResult.Data.(*[]gwml.Config), nil
}

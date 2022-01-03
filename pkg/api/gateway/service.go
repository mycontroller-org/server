package gateway

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/type"
	"go.uber.org/zap"
)

// Start gateway
func Start(gwCfg *gwTY.Config) error {
	return postGatewayCommand(gwCfg, rsTY.CommandStart)
}

// Stop gateway
func Stop(gwCfg *gwTY.Config) error {
	return postGatewayCommand(gwCfg, rsTY.CommandStop)
}

// LoadAll makes gateways alive
func LoadAll() {
	gwsResult, err := List(nil, nil)
	if err != nil {
		zap.L().Error("Failed to get list of gateways", zap.Error(err))
		return
	}
	gateways := *gwsResult.Data.(*[]gwTY.Config)
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
	err := postGatewayCommand(nil, rsTY.CommandUnloadAll)
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

func postGatewayCommand(gwCfg *gwTY.Config, command string) error {
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeGateway,
		Command: command,
	}
	if gwCfg != nil {
		reqEvent.ID = gwCfg.ID
		reqEvent.SetData(gwCfg)
	}
	topic := mcbus.FormatTopic(mcbus.TopicServiceGateway)
	return mcbus.Publish(topic, reqEvent)
}

func getGatewayEntries(ids []string) ([]gwTY.Config, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	gwsResult, err := List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *gwsResult.Data.(*[]gwTY.Config), nil
}

package gateway

import (
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

// Start gateway
func (gw *GatewayAPI) Start(gwCfg *gwTY.Config) error {
	return gw.postGatewayCommand(gwCfg, rsTY.CommandStart)
}

// Stop gateway
func (gw *GatewayAPI) Stop(gwCfg *gwTY.Config) error {
	return gw.postGatewayCommand(gwCfg, rsTY.CommandStop)
}

// LoadAll makes gateways alive
func (gw *GatewayAPI) LoadAll() {
	gwsResult, err := gw.List(nil, nil)
	if err != nil {
		gw.logger.Error("failed to get list of gateways", zap.Error(err))
		return
	}
	gateways := *gwsResult.Data.(*[]gwTY.Config)
	for index := 0; index < len(gateways); index++ {
		gateway := gateways[index]
		if gateway.Enabled {
			err = gw.Start(&gateway)
			if err != nil {
				gw.logger.Error("failed to load a gateway", zap.Error(err), zap.String("gateway", gateway.ID))
			}
		}
	}
}

// UnloadAll makes stop all gateways
func (gw *GatewayAPI) UnloadAll() {
	err := gw.postGatewayCommand(nil, rsTY.CommandUnloadAll)
	if err != nil {
		gw.logger.Error("error on unload gateways command", zap.Error(err))
	}
}

// Enable gateway
func (gw *GatewayAPI) Enable(ids []string) error {
	gateways, err := gw.getGatewayEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(gateways); index++ {
		gwCfg := gateways[index]
		if !gwCfg.Enabled {
			gwCfg.Enabled = true
			err = gw.Save(&gwCfg)
			if err != nil {
				return err
			}
			return gw.Start(&gwCfg)
		}
	}
	return nil
}

// Disable gateway
func (gw *GatewayAPI) Disable(ids []string) error {
	gateways, err := gw.getGatewayEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(gateways); index++ {
		gwCfg := gateways[index]
		if gwCfg.Enabled {
			gwCfg.Enabled = false
			err = gw.Save(&gwCfg)
			if err != nil {
				return err
			}
			return gw.Stop(&gwCfg)
		}
	}
	return nil
}

// Reload gateway
func (gw *GatewayAPI) Reload(ids []string) error {
	gateways, err := gw.getGatewayEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(gateways); index++ {
		gateway := gateways[index]
		err = gw.Stop(&gateway)
		if err != nil {
			gw.logger.Error("error on stopping a gateway command", zap.Error(err), zap.String("gateway", gateway.ID))
		}
		if gateway.Enabled {
			err = gw.Start(&gateway)
			if err != nil {
				gw.logger.Error("error on start a gateway command", zap.Error(err), zap.String("gateway", gateway.ID))
			}
		}
	}
	return nil
}

func (gw *GatewayAPI) postGatewayCommand(gwCfg *gwTY.Config, command string) error {
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeGateway,
		Command: command,
	}
	if gwCfg != nil {
		reqEvent.ID = gwCfg.ID
		reqEvent.SetData(gwCfg)
	}
	return gw.bus.Publish(topic.TopicServiceGateway, reqEvent)
}

func (gw *GatewayAPI) getGatewayEntries(ids []string) ([]gwTY.Config, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	gwsResult, err := gw.List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *gwsResult.Data.(*[]gwTY.Config), nil
}

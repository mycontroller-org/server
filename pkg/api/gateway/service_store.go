package gateway

import (
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	gwpd "github.com/mycontroller-org/backend/v2/plugin/gw_provider"
)

var gwService = map[string]*gwpd.Service{}

// AddGatewayService add
func AddGatewayService(service *gwpd.Service) {
	gwService[service.GatewayConfig.ID] = service
}

// RemoveGatewayService remove a service
func RemoveGatewayService(gatewayCfg *gwml.Config) {
	delete(gwService, gatewayCfg.ID)
}

// GetGatewayService returns service
func GetGatewayService(gatewayCfg *gwml.Config) *gwpd.Service {
	return GetGatewayServiceByID(gatewayCfg.ID)
}

// GetGatewayServiceByID returns service
func GetGatewayServiceByID(ID string) *gwpd.Service {
	return gwService[ID]
}

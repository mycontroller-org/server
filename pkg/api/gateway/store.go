package gateway

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/mcbus"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
)

var gwService = map[string]*Service{}

// AddGatewayService add
func AddGatewayService(service *Service) {
	gwService[service.Config.ID] = service
}

// RemoveGatewayService remove a service
func RemoveGatewayService(gatewayCfg *gwml.Config) {
	delete(gwService, gatewayCfg.ID)
}

// GetGatewayService returns service
func GetGatewayService(gatewayCfg *gwml.Config) *Service {
	return GetGatewayServiceByID(gatewayCfg.ID)
}

// GetGatewayServiceByID returns service
func GetGatewayServiceByID(ID string) *Service {
	return gwService[ID]
}

// Post a message to gateway
func Post(msg *msgml.Message) error {
	gatewayService := GetGatewayServiceByID(msg.GatewayID)
	if gatewayService == nil {
		return fmt.Errorf("Gateway service not found for %s", msg.GatewayID)
	}
	topic := gatewayService.TopicMsg2Provider
	_, err := mcbus.Publish(topic, msg)
	return err
}

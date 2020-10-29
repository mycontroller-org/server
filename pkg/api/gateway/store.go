package gateway

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/mcbus"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
)

var gwService = map[string]*Service{}

// AddGatewayService add
func AddGatewayService(s *Service) {
	gwService[s.Config.ID] = s
}

// RemoveGatewayService remove a service
func RemoveGatewayService(g *gwml.Config) {
	delete(gwService, g.ID)
}

// GetGatewayService returns service
func GetGatewayService(g *gwml.Config) *Service {
	return GetGatewayServiceByID(g.ID)
}

// GetGatewayServiceByID returns service
func GetGatewayServiceByID(ID string) *Service {
	return gwService[ID]
}

// Post a message to gateway
func Post(msg *msgml.Message) error {
	gwSRV := GetGatewayServiceByID(msg.GatewayID)
	if gwSRV == nil {
		return fmt.Errorf("Gateway service not found for %s", msg.GatewayID)
	}
	topic := gwSRV.TopicMsg2Provider
	_, err := mcbus.Publish(topic, msg)
	return err
}

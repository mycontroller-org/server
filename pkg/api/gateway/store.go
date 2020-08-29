package gateway

import (
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
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
	return gwService[g.ID]
}

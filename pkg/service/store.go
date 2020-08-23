package service

import (
	gwml "github.com/mycontroller-org/backend/pkg/model/gateway"
)

var gwService = map[string]*gwml.Service{}

// AddGatewayService add
func AddGatewayService(s *gwml.Service) {
	gwService[s.Config.ID] = s
}

// RemoveGatewayService remove a service
func RemoveGatewayService(g *gwml.Config) {
	delete(gwService, g.ID)
}

// GetGatewayService returns service
func GetGatewayService(g *gwml.Config) *gwml.Service {
	return gwService[g.ID]
}

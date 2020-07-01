package service

import (
	ml "github.com/mycontroller-org/mycontroller/pkg/model"
)

var gwService = map[string]*ml.GatewayService{}

// AddGatewayService add
func AddGatewayService(s *ml.GatewayService) {
	gwService[s.Config.ID] = s
}

// RemoveGatewayService remove a service
func RemoveGatewayService(g *ml.GatewayConfig) {
	delete(gwService, g.ID)
}

// GetGatewayService returns service
func GetGatewayService(g *ml.GatewayConfig) *ml.GatewayService {
	return gwService[g.ID]
}

package gateway

import (
	"sync"

	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	gwpd "github.com/mycontroller-org/backend/v2/plugin/gw_provider"
)

type gatewayService struct {
	services map[string]*gwpd.Service
	mutex    sync.Mutex
}

var gwService = gatewayService{
	services: make(map[string]*gwpd.Service),
}

// Add a service
func (gs *gatewayService) Add(service *gwpd.Service) {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	gs.services[service.GatewayConfig.ID] = service
}

// Remove a service
func (gs *gatewayService) Remove(gatewayCfg *gwml.Config) {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	delete(gs.services, gatewayCfg.ID)
}

// Get returns a service
func (gs *gatewayService) Get(gatewayCfg *gwml.Config) *gwpd.Service {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	return gs.services[gatewayCfg.ID]
}

// GetByID returns service by id
func (gs *gatewayService) GetByID(ID string) *gwpd.Service {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	return gs.services[ID]
}

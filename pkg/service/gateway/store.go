package service

import (
	"sync"

	gwProvider "github.com/mycontroller-org/server/v2/plugin/gateway/provider"
)

type Store struct {
	services map[string]*gwProvider.Service
	mutex    sync.Mutex
}

// Add a service
func (s *Store) Add(service *gwProvider.Service) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.services[service.GatewayConfig.ID] = service
}

// Remove a service
func (s *Store) Remove(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.services, id)
}

// GetByID returns service by id
func (s *Store) Get(id string) *gwProvider.Service {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if service, found := s.services[id]; found {
		return service
	}
	return nil
}

func (s *Store) ListIDs() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ids := make([]string, 0)
	for id := range s.services {
		ids = append(ids, id)
	}
	return ids
}

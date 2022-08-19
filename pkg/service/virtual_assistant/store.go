package service

import (
	"sync"

	vaTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_assistant"
)

type store struct {
	services map[string]vaTY.Plugin
	mutex    sync.Mutex
}

var vaService = store{
	services: make(map[string]vaTY.Plugin),
}

// Add a service
func (s *store) Add(service vaTY.Plugin) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.services[service.Config().ID] = service
}

// Remove a service
func (s *store) Remove(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.services, id)
}

// GetByID returns service by id
func (s *store) Get(id string) vaTY.Plugin {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if service, found := s.services[id]; found {
		return service
	}
	return nil
}

func (s *store) ListIDs() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ids := make([]string, 0)
	for id := range s.services {
		ids = append(ids, id)
	}
	return ids
}

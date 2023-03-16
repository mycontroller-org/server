package service

import (
	"sync"

	vaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/types"
)

type Store struct {
	services map[string]vaTY.Plugin
	mutex    sync.Mutex
}

// Add a service
func (s *Store) Add(service vaTY.Plugin) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.services[service.Config().ID] = service
}

// Remove a service
func (s *Store) Remove(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.services, id)
}

// GetByID returns service by id
func (s *Store) Get(id string) vaTY.Plugin {
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

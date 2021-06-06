package esphome

import (
	"sync"
)

// EntityStore keeps all the esphome node available entities with key and sourceID
type EntityStore struct {
	nodes map[string]map[uint32]Entity
	mutex *sync.RWMutex
}

// AddEntity adds a entity to the store
func (s *EntityStore) AddEntity(nodeID string, key uint32, entity Entity) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, found := s.nodes[nodeID]; !found {
		s.nodes[nodeID] = make(map[uint32]Entity)
	}
	s.nodes[nodeID][key] = *entity.Clone()
}

// GetByKey returns a entity by a key
func (s *EntityStore) GetByKey(nodeID string, key uint32) *Entity {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if entityMap, found := s.nodes[nodeID]; found {
		if entity, keyFound := entityMap[key]; keyFound {
			return entity.Clone()
		}
	}
	return nil
}

// GetBySourceID returns a entity by sourceID
func (s *EntityStore) GetBySourceID(nodeID string, sourceID string) *Entity {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if entityMap, found := s.nodes[nodeID]; found {
		for _, entity := range entityMap {
			if entity.SourceID == sourceID {
				return entity.Clone()
			}
		}
	}
	return nil
}

// GetByEntityType returns a entity by entity type
func (s *EntityStore) GetByEntityType(nodeID string, entityType string) *Entity {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if entityMap, found := s.nodes[nodeID]; found {
		for _, entity := range entityMap {
			if entity.Type == entityType {
				return entity.Clone()
			}
		}
	}
	return nil
}

func (s *EntityStore) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.nodes = make(map[string]map[uint32]Entity)
}

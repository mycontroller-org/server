package concurrency

import (
	"sync"
)

// Store struct
type Store struct {
	Name  string
	data  map[string]interface{}
	mutex sync.RWMutex
}

// NewStore returns brandnew store
func NewStore() *Store {
	return &Store{data: make(map[string]interface{})}
}

// Add a value to the store
func (s *Store) Add(key string, value interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = value
}

// Get a value from the store
func (s *Store) Get(key string) interface{} {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, ok := s.data[key]
	if !ok {
		return nil
	}
	return value
}

// Remove a value to the store
func (s *Store) Remove(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.data, key)
}

// IsAvailable returns the availability
func (s *Store) IsAvailable(key string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, found := s.data[key]
	return found
}

// Keys returns the available keys
func (s *Store) Keys() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	keys := make([]string, 0)
	for key := range s.data {
		keys = append(keys, key)
	}
	return keys
}

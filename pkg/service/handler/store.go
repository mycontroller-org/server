package handler

import (
	"sync"

	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

type Store struct {
	handlers map[string]handlerTY.Plugin
	mutex    sync.Mutex
	logger   *zap.Logger
}

// Add a handler
func (s *Store) Add(id string, handler handlerTY.Plugin) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.handlers[id] = handler
}

// Remove a handler
func (s *Store) Remove(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.handlers, id)
}

// GetByID returns handler by id
func (s *Store) Get(id string) handlerTY.Plugin {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.handlers[id]
}

func (s *Store) CloseHandlers() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for id := range s.handlers {
		handler := s.handlers[id]
		err := handler.Close()
		if err != nil {
			s.logger.Error("error on close a handler", zap.String("id", id), zap.Error(err))
		}
	}
	s.handlers = make(map[string]handlerTY.Plugin)
}

func (s *Store) ListIDs() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ids := make([]string, 0)
	for id := range s.handlers {
		ids = append(ids, id)
	}
	return ids
}

package handler

import (
	"sync"

	handler "github.com/mycontroller-org/backend/v2/plugin/handlers"
	"go.uber.org/zap"
)

type store struct {
	handlers map[string]handler.Handler
	mutex    sync.Mutex
}

var handlersStore = store{
	handlers: make(map[string]handler.Handler),
}

// Add a handler
func (s *store) Add(id string, handler handler.Handler) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.handlers[id] = handler
}

// Remove a handler
func (s *store) Remove(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.handlers, id)
}

// GetByID returns handler by id
func (s *store) Get(id string) handler.Handler {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.handlers[id]
}

func (s *store) closeHandlers() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for id := range handlersStore.handlers {
		handler := handlersStore.handlers[id]
		err := handler.Close()
		if err != nil {
			zap.L().Error("error on close a handler", zap.String("id", id), zap.Error(err))
		}
	}
	handlersStore.handlers = make(map[string]handler.Handler)
}

func (s *store) ListIDs() []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ids := make([]string, 0)
	for id := range s.handlers {
		ids = append(ids, id)
	}
	return ids
}

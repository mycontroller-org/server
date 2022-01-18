package handler

import (
	"sync"

	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

type store struct {
	handlers map[string]handlerTY.Plugin
	mutex    sync.Mutex
}

var handlersStore = store{
	handlers: make(map[string]handlerTY.Plugin),
}

// Add a handler
func (s *store) Add(id string, handler handlerTY.Plugin) {
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
func (s *store) Get(id string) handlerTY.Plugin {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.handlers[id]
}

func (s *store) CloseHandlers() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for id := range handlersStore.handlers {
		handler := handlersStore.handlers[id]
		err := handler.Close()
		if err != nil {
			zap.L().Error("error on close a handler", zap.String("id", id), zap.Error(err))
		}
	}
	handlersStore.handlers = make(map[string]handlerTY.Plugin)
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

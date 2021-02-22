package handler

import (
	"fmt"

	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	"github.com/mycontroller-org/backend/v2/plugin/notify_handlers/email"
	"github.com/mycontroller-org/backend/v2/plugin/notify_handlers/noop"
)

// Handler interface details for operation
type Handler interface {
	Start() error
	Close() error
	Post(variables map[string]interface{}) error
}

// GetHandler loads and returns a handler
func GetHandler(cfg *handlerML.Config) (Handler, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("handler disabled. id: %s", cfg.ID)
	}

	switch cfg.Type {
	case handlerML.TypeEmail:
		return email.Init(cfg.ID, cfg.Spec)

	case handlerML.TypeNoop:
		return &noop.Client{}, nil

	default:
		return nil, fmt.Errorf("unsupported handler, id:%s, name:%s, type:%s", cfg.ID, cfg.Description, cfg.Type)
	}
}

package handler

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	"github.com/mycontroller-org/backend/v2/plugin/notify_handlers/email"
	"github.com/mycontroller-org/backend/v2/plugin/notify_handlers/noop"
	resourceAction "github.com/mycontroller-org/backend/v2/plugin/notify_handlers/resource"
)

// Handler interface details for operation
type Handler interface {
	Start() error
	Close() error
	Post(variables map[string]interface{}) error
	State() *model.State
}

// GetHandler loads and returns a handler
func GetHandler(cfg *handlerML.Config) (Handler, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("handler disabled. id: %s", cfg.ID)
	}

	switch cfg.Type {
	case handlerML.TypeEmail:
		return email.Init(cfg)

	case handlerML.TypeNoop:
		return &noop.Client{HandlerCfg: cfg}, nil

	case handlerML.TypeResource:
		return &resourceAction.Client{HandlerCfg: cfg}, nil

	default:
		return nil, fmt.Errorf("unsupported handler, id:%s, name:%s, type:%s", cfg.ID, cfg.Description, cfg.Type)
	}
}

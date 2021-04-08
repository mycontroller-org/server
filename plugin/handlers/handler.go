package handler

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	"github.com/mycontroller-org/backend/v2/plugin/handlers/email"
	"github.com/mycontroller-org/backend/v2/plugin/handlers/exporter"
	"github.com/mycontroller-org/backend/v2/plugin/handlers/noop"
	resourceAction "github.com/mycontroller-org/backend/v2/plugin/handlers/resource"
	"github.com/mycontroller-org/backend/v2/plugin/handlers/telegram"
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

	case handlerML.TypeTelegram:
		return telegram.Init(cfg)

	case handlerML.TypeExporter:
		return exporter.Init(cfg)

	default:
		return nil, fmt.Errorf("unsupported handler, id:%s, name:%s, type:%s", cfg.ID, cfg.Description, cfg.Type)
	}
}

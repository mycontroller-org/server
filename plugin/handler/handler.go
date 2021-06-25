package handler

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/model"
	handlerML "github.com/mycontroller-org/server/v2/pkg/model/handler"
	backupHandler "github.com/mycontroller-org/server/v2/plugin/handler/backup"
	emailHandler "github.com/mycontroller-org/server/v2/plugin/handler/email"
	noopHandler "github.com/mycontroller-org/server/v2/plugin/handler/noop"
	resourceAction "github.com/mycontroller-org/server/v2/plugin/handler/resource"
	telegramHandler "github.com/mycontroller-org/server/v2/plugin/handler/telegram"
	webhookHandler "github.com/mycontroller-org/server/v2/plugin/handler/webhook"
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
		return emailHandler.Init(cfg)

	case handlerML.TypeNoop:
		return &noopHandler.Client{HandlerCfg: cfg}, nil

	case handlerML.TypeResource:
		return &resourceAction.Client{HandlerCfg: cfg}, nil

	case handlerML.TypeTelegram:
		return telegramHandler.Init(cfg)

	case handlerML.TypeWebhook:
		return webhookHandler.Init(cfg)

	case handlerML.TypeBackup:
		return backupHandler.Init(cfg)

	default:
		return nil, fmt.Errorf("unsupported handler, id:%s, name:%s, type:%s", cfg.ID, cfg.Description, cfg.Type)
	}
}

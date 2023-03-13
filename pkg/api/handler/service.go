package handler

import (
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// Start notifyHandler
func (h *HandlerAPI) Start(cfg *handlerTY.Config) error {
	return h.postCommand(cfg, rsTY.CommandStart)
}

// Stop notifyHandler
func (h *HandlerAPI) Stop(cfg *handlerTY.Config) error {
	return h.postCommand(cfg, rsTY.CommandStop)
}

// LoadAll makes notifyHandlers alive
func (h *HandlerAPI) LoadAll() {
	result, err := h.List(nil, nil)
	if err != nil {
		h.logger.Error("Failed to get list of handlers", zap.Error(err))
		return
	}
	handlers := *result.Data.(*[]handlerTY.Config)
	for index := 0; index < len(handlers); index++ {
		cfg := handlers[index]
		if cfg.Enabled {
			err = h.Start(&cfg)
			if err != nil {
				h.logger.Error("error on load a handler", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
}

// UnloadAll makes stop all notifyHandlers
func (h *HandlerAPI) UnloadAll() {
	err := h.postCommand(nil, rsTY.CommandUnloadAll)
	if err != nil {
		h.logger.Error("error on unloadall handlers command", zap.Error(err))
	}
}

// Enable notifyHandler
func (h *HandlerAPI) Enable(ids []string) error {
	notifyHandlers, err := h.getNotifyHandlerEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(notifyHandlers); index++ {
		cfg := notifyHandlers[index]
		if !cfg.Enabled {
			cfg.Enabled = true
			err = h.SaveAndReload(&cfg)
			if err != nil {
				h.logger.Error("error on enabling a handler", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
	return nil
}

// Disable notifyHandler
func (h *HandlerAPI) Disable(ids []string) error {
	notifyHandlers, err := h.getNotifyHandlerEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(notifyHandlers); index++ {
		cfg := notifyHandlers[index]
		err := h.Stop(&cfg)
		if err != nil {
			h.logger.Error("error on disabling a handler", zap.Error(err), zap.String("id", cfg.ID))
		}
		if cfg.Enabled {
			cfg.Enabled = false
			err = h.Save(&cfg)
			if err != nil {
				h.logger.Error("error on saving a handler", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
	return nil
}

// Reload notifyHandler
func (h *HandlerAPI) Reload(ids []string) error {
	notifyHandlers, err := h.getNotifyHandlerEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(notifyHandlers); index++ {
		cfg := notifyHandlers[index]
		if cfg.Enabled {
			err = h.postCommand(&cfg, rsTY.CommandReload)
			if err != nil {
				h.logger.Error("error on reload handler command", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
	return nil
}

func (h *HandlerAPI) postCommand(cfg *handlerTY.Config, command string) error {
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeHandler,
		Command: command,
	}
	if cfg != nil {
		reqEvent.ID = cfg.ID
		reqEvent.SetData(cfg)
	}
	return h.bus.Publish(topic.TopicServiceHandler, reqEvent)
}

func (h *HandlerAPI) getNotifyHandlerEntries(ids []string) ([]handlerTY.Config, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	result, err := h.List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *result.Data.(*[]handlerTY.Config), nil
}

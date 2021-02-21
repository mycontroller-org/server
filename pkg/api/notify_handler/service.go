package handler

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	rsml "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

// Start notifyHandler
func Start(cfg *handlerML.Config) error {
	return postCommand(cfg, rsml.CommandStart)
}

// Stop notifyHandler
func Stop(cfg *handlerML.Config) error {
	return postCommand(cfg, rsml.CommandStop)
}

// LoadAll makes notifyHandlers alive
func LoadAll() {
	result, err := List(nil, nil)
	if err != nil {
		zap.L().Error("Failed to get list of notifyHandlers", zap.Error(err))
		return
	}
	notifyHandlers := *result.Data.(*[]handlerML.Config)
	for index := 0; index < len(notifyHandlers); index++ {
		notifyHandler := notifyHandlers[index]
		if notifyHandler.Enabled {
			err = Start(&notifyHandler)
			if err != nil {
				zap.L().Error("Failed to load a notifyHandler", zap.Error(err), zap.String("notifyHandler", notifyHandler.ID))
			}
		}
	}
}

// UnloadAll makes stop all notifyHandlers
func UnloadAll() {
	err := postCommand(nil, rsml.CommandUnloadAll)
	if err != nil {
		zap.L().Error("error on unload notifyHandlers command", zap.Error(err))
	}
}

// Enable notifyHandler
func Enable(ids []string) error {
	notifyHandlers, err := getNotifyHandlerEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(notifyHandlers); index++ {
		cfg := notifyHandlers[index]
		if !cfg.Enabled {
			cfg.Enabled = true
			err = Save(&cfg)
			if err != nil {
				return err
			}
			return postCommand(&cfg, rsml.CommandStart)
		}
	}
	return nil
}

// Disable notifyHandler
func Disable(ids []string) error {
	notifyHandlers, err := getNotifyHandlerEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(notifyHandlers); index++ {
		cfg := notifyHandlers[index]
		if cfg.Enabled {
			cfg.Enabled = false
			err = Save(&cfg)
			if err != nil {
				return err
			}
			err = postCommand(&cfg, rsml.CommandStop)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Reload notifyHandler
func Reload(ids []string) error {
	notifyHandlers, err := getNotifyHandlerEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(notifyHandlers); index++ {
		notifyHandler := notifyHandlers[index]
		if notifyHandler.Enabled {
			err = postCommand(&notifyHandler, rsml.CommandReload)
			if err != nil {
				zap.L().Error("error on posting notifyHandler reload command", zap.Error(err), zap.String("notifyHandler", notifyHandler.ID))
			}
		}
	}
	return nil
}

func postCommand(cfg *handlerML.Config, command string) error {
	reqEvent := rsml.Event{
		Type:    rsml.TypeNotifyHandler,
		Command: command,
	}
	if cfg != nil {
		reqEvent.ID = cfg.ID
		err := reqEvent.SetData(cfg)
		if err != nil {
			return err
		}
	}
	topic := mcbus.FormatTopic(mcbus.TopicServiceNotifyHandler)
	return mcbus.Publish(topic, reqEvent)
}

func getNotifyHandlerEntries(ids []string) ([]handlerML.Config, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: ids}}
	pagination := &stgml.Pagination{Limit: 100}
	result, err := List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *result.Data.(*[]handlerML.Config), nil
}

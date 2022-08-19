package assistant

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	vaTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_assistant"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// Add virtual assistant
func Add(cfg *vaTY.Config) error {
	return postCommand(cfg, rsTY.CommandAdd)
}

// Remove virtual assistant
func Remove(cfg *vaTY.Config) error {
	return postCommand(cfg, rsTY.CommandRemove)
}

// LoadAll makes virtual assistants alive
func LoadAll() {
	result, err := List(nil, nil)
	if err != nil {
		zap.L().Error("failed to get list of virtual assistants", zap.Error(err))
		return
	}
	vas := *result.Data.(*[]vaTY.Config)
	for index := 0; index < len(vas); index++ {
		cfg := vas[index]
		if cfg.Enabled {
			err = Add(&cfg)
			if err != nil {
				zap.L().Error("failed to load a virtual assistant", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
}

// UnloadAll makes stop all virtual assistants
func UnloadAll() {
	err := postCommand(nil, rsTY.CommandUnloadAll)
	if err != nil {
		zap.L().Error("error on unloadall virtual assistant command", zap.Error(err))
	}
}

// Enable virtual assistant
func Enable(ids []string) error {
	vas, err := getVirtualAssistantEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(vas); index++ {
		cfg := vas[index]
		if !cfg.Enabled {
			cfg.Enabled = true
			err = SaveAndReload(&cfg)
			if err != nil {
				zap.L().Error("error on enabling a virtual assistant", zap.String("id", cfg.ID), zap.Error(err))
			}
		}
	}
	return nil
}

// Disable virtual assistant
func Disable(ids []string) error {
	vas, err := getVirtualAssistantEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(vas); index++ {
		cfg := vas[index]
		if cfg.Enabled {
			cfg.Enabled = false
			err = Save(&cfg)
			if err != nil {
				return err
			}
			err = Remove(&cfg)
			if err != nil {
				zap.L().Error("error on disabling a virtual assistant", zap.String("id", cfg.ID), zap.Error(err))
			}
		}
	}
	return nil
}

// Reload virtual assistant
func Reload(ids []string) error {
	vas, err := getVirtualAssistantEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(vas); index++ {
		cfg := vas[index]
		err = Remove(&cfg)
		if err != nil {
			zap.L().Error("error on removing a virtual assistant", zap.Error(err), zap.String("id", cfg.ID))
		}
		if cfg.Enabled {
			err = Add(&cfg)
			if err != nil {
				zap.L().Error("error on adding a virtual assistant", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
	return nil
}

func postCommand(cfg *vaTY.Config, command string) error {
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeVirtualAssistant,
		Command: command,
	}
	if cfg != nil {
		reqEvent.ID = cfg.ID
		reqEvent.SetData(cfg)
	}
	topic := mcbus.FormatTopic(mcbus.TopicServiceVirtualAssistant)
	return mcbus.Publish(topic, reqEvent)
}

func getVirtualAssistantEntries(ids []string) ([]vaTY.Config, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	result, err := List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *result.Data.(*[]vaTY.Config), nil
}

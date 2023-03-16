package assistant

import (
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	vaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// Add virtual assistant
func (va *VirtualAssistantAPI) Add(cfg *vaTY.Config) error {
	return va.postCommand(cfg, rsTY.CommandAdd)
}

// Remove virtual assistant
func (va *VirtualAssistantAPI) Remove(cfg *vaTY.Config) error {
	return va.postCommand(cfg, rsTY.CommandRemove)
}

// LoadAll makes virtual assistants alive
func (va *VirtualAssistantAPI) LoadAll() {
	result, err := va.List(nil, nil)
	if err != nil {
		va.logger.Error("failed to get list of virtual assistants", zap.Error(err))
		return
	}
	vas := *result.Data.(*[]vaTY.Config)
	for index := 0; index < len(vas); index++ {
		cfg := vas[index]
		if cfg.Enabled {
			err = va.Add(&cfg)
			if err != nil {
				va.logger.Error("failed to load a virtual assistant", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
}

// UnloadAll makes stop all virtual assistants
func (va *VirtualAssistantAPI) UnloadAll() {
	err := va.postCommand(nil, rsTY.CommandUnloadAll)
	if err != nil {
		va.logger.Error("error on unloadall virtual assistant command", zap.Error(err))
	}
}

// Enable virtual assistant
func (va *VirtualAssistantAPI) Enable(ids []string) error {
	vas, err := va.getVirtualAssistantEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(vas); index++ {
		cfg := vas[index]
		if !cfg.Enabled {
			cfg.Enabled = true
			err = va.SaveAndReload(&cfg)
			if err != nil {
				va.logger.Error("error on enabling a virtual assistant", zap.String("id", cfg.ID), zap.Error(err))
			}
		}
	}
	return nil
}

// Disable virtual assistant
func (va *VirtualAssistantAPI) Disable(ids []string) error {
	vas, err := va.getVirtualAssistantEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(vas); index++ {
		cfg := vas[index]
		if cfg.Enabled {
			cfg.Enabled = false
			err = va.Save(&cfg)
			if err != nil {
				return err
			}
			err = va.Remove(&cfg)
			if err != nil {
				va.logger.Error("error on disabling a virtual assistant", zap.String("id", cfg.ID), zap.Error(err))
			}
		}
	}
	return nil
}

// Reload virtual assistant
func (va *VirtualAssistantAPI) Reload(ids []string) error {
	vas, err := va.getVirtualAssistantEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(vas); index++ {
		cfg := vas[index]
		err = va.Remove(&cfg)
		if err != nil {
			va.logger.Error("error on removing a virtual assistant", zap.Error(err), zap.String("id", cfg.ID))
		}
		if cfg.Enabled {
			err = va.Add(&cfg)
			if err != nil {
				va.logger.Error("error on adding a virtual assistant", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
	return nil
}

func (va *VirtualAssistantAPI) postCommand(cfg *vaTY.Config, command string) error {
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeVirtualAssistant,
		Command: command,
	}
	if cfg != nil {
		reqEvent.ID = cfg.ID
		reqEvent.SetData(cfg)
	}
	return va.bus.Publish(topic.TopicServiceVirtualAssistant, reqEvent)
}

func (va *VirtualAssistantAPI) getVirtualAssistantEntries(ids []string) ([]vaTY.Config, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	result, err := va.List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *result.Data.(*[]vaTY.Config), nil
}

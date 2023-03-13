package schedule

import (
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// Add scheduler
func (sh *ScheduleAPI) Add(cfg *schedulerTY.Config) error {
	return sh.postCommand(cfg, rsTY.CommandAdd)
}

// Remove scheduler
func (sh *ScheduleAPI) Remove(cfg *schedulerTY.Config) error {
	return sh.postCommand(cfg, rsTY.CommandRemove)
}

// LoadAll makes schedulers alive
func (sh *ScheduleAPI) LoadAll() {
	result, err := sh.List(nil, nil)
	if err != nil {
		sh.logger.Error("failed to get list of schedules", zap.Error(err))
		return
	}
	schedulers := *result.Data.(*[]schedulerTY.Config)
	for index := 0; index < len(schedulers); index++ {
		cfg := schedulers[index]
		if cfg.Enabled {
			err = sh.Add(&cfg)
			if err != nil {
				sh.logger.Error("failed to load a schedule", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
}

// UnloadAll makes stop all schedulers
func (sh *ScheduleAPI) UnloadAll() {
	err := sh.postCommand(nil, rsTY.CommandUnloadAll)
	if err != nil {
		sh.logger.Error("error on unloadall scheduler command", zap.Error(err))
	}
}

// Enable scheduler
func (sh *ScheduleAPI) Enable(ids []string) error {
	schedulers, err := sh.getSchedulerEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(schedulers); index++ {
		cfg := schedulers[index]
		if !cfg.Enabled {
			cfg.Enabled = true
			err = sh.SaveAndReload(&cfg)
			if err != nil {
				sh.logger.Error("error on enabling a schedule", zap.String("id", cfg.ID), zap.Error(err))
			}
		}
	}
	return nil
}

// Disable scheduler
func (sh *ScheduleAPI) Disable(ids []string) error {
	schedulers, err := sh.getSchedulerEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(schedulers); index++ {
		cfg := schedulers[index]
		if cfg.Enabled {
			cfg.Enabled = false
			err = sh.Save(&cfg)
			if err != nil {
				return err
			}
			err = sh.Remove(&cfg)
			if err != nil {
				sh.logger.Error("error on disabling a schedule", zap.String("id", cfg.ID), zap.Error(err))
			}
		}
	}
	return nil
}

// Reload scheduler
func (sh *ScheduleAPI) Reload(ids []string) error {
	schedules, err := sh.getSchedulerEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(schedules); index++ {
		cfg := schedules[index]
		err = sh.Remove(&cfg)
		if err != nil {
			sh.logger.Error("error on removing a scheduling", zap.Error(err), zap.String("id", cfg.ID))
		}
		if cfg.Enabled {
			err = sh.Add(&cfg)
			if err != nil {
				sh.logger.Error("error on adding a schedule", zap.Error(err), zap.String("id", cfg.ID))
			}
		}
	}
	return nil
}

func (sh *ScheduleAPI) postCommand(cfg *schedulerTY.Config, command string) error {
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeScheduler,
		Command: command,
	}
	if cfg != nil {
		reqEvent.ID = cfg.ID
		reqEvent.SetData(cfg)
	}
	return sh.bus.Publish(topic.TopicServiceScheduler, reqEvent)
}

func (sh *ScheduleAPI) getSchedulerEntries(ids []string) ([]schedulerTY.Config, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	result, err := sh.List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *result.Data.(*[]schedulerTY.Config), nil
}

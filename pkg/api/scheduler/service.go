package scheduler

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	schedulerML "github.com/mycontroller-org/backend/v2/pkg/model/scheduler"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

// Add scheduler
func Add(cfg *schedulerML.Config) error {
	return postCommand(cfg, rsML.CommandAdd)
}

// Remove scheduler
func Remove(cfg *schedulerML.Config) error {
	return postCommand(cfg, rsML.CommandRemove)
}

// LoadAll makes schedulers alive
func LoadAll() {
	result, err := List(nil, nil)
	if err != nil {
		zap.L().Error("Failed to get list of schedules", zap.Error(err))
		return
	}
	schedulers := *result.Data.(*[]schedulerML.Config)
	for index := 0; index < len(schedulers); index++ {
		scheduler := schedulers[index]
		if scheduler.Enabled {
			err = Add(&scheduler)
			if err != nil {
				zap.L().Error("Failed to load a schedule", zap.Error(err), zap.String("scheduleID", scheduler.ID))
			}
		}
	}
}

// UnloadAll makes stop all schedulers
func UnloadAll() {
	err := postCommand(nil, rsML.CommandUnloadAll)
	if err != nil {
		zap.L().Error("error on unloadall scheduler command", zap.Error(err))
	}
}

// Enable scheduler
func Enable(ids []string) error {
	schedulers, err := getSchedulerEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(schedulers); index++ {
		cfg := schedulers[index]
		if !cfg.Enabled {
			cfg.Enabled = true
			err = Save(&cfg)
			if err != nil {
				return err
			}
			return postCommand(&cfg, rsML.CommandAdd)
		}
	}
	return nil
}

// Disable scheduler
func Disable(ids []string) error {
	schedulers, err := getSchedulerEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(schedulers); index++ {
		cfg := schedulers[index]
		if cfg.Enabled {
			cfg.Enabled = false
			err = Save(&cfg)
			if err != nil {
				return err
			}
			err = postCommand(&cfg, rsML.CommandRemove)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Reload scheduler
func Reload(ids []string) error {
	schedules, err := getSchedulerEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(schedules); index++ {
		schedule := schedules[index]
		err = Remove(&schedule)
		if err != nil {
			zap.L().Error("error on posting scheduler remove command", zap.Error(err), zap.String("scheduleID", schedule.ID))
		}
		if schedule.Enabled {
			err = Add(&schedule)
			if err != nil {
				zap.L().Error("error on posting scheduler add command", zap.Error(err), zap.String("scheduleID", schedule.ID))
			}
		}
	}
	return nil
}

func postCommand(cfg *schedulerML.Config, command string) error {
	reqEvent := rsML.Event{
		Type:    rsML.TypeScheduler,
		Command: command,
	}
	if cfg != nil {
		reqEvent.ID = cfg.ID
		err := reqEvent.SetData(cfg)
		if err != nil {
			return err
		}
	}
	topic := mcbus.FormatTopic(mcbus.TopicServiceScheduler)
	return mcbus.Publish(topic, reqEvent)
}

func getSchedulerEntries(ids []string) ([]schedulerML.Config, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: ids}}
	pagination := &stgml.Pagination{Limit: 100}
	result, err := List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *result.Data.(*[]schedulerML.Config), nil
}

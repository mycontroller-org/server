package task

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// Add task
func Add(cfg *taskTY.Config) error {
	return postCommand(cfg, rsTY.CommandAdd)
}

// Remove task
func Remove(cfg *taskTY.Config) error {
	return postCommand(cfg, rsTY.CommandRemove)
}

// LoadAll makes tasks alive
func LoadAll() {
	result, err := List(nil, nil)
	if err != nil {
		zap.L().Error("Failed to get list of tasks", zap.Error(err))
		return
	}
	tasks := *result.Data.(*[]taskTY.Config)
	for index := 0; index < len(tasks); index++ {
		task := tasks[index]
		if task.Enabled {
			err = Add(&task)
			if err != nil {
				zap.L().Error("Failed to load a task", zap.Error(err), zap.String("task", task.ID))
			}
		}
	}
}

// UnloadAll makes stop all tasks
func UnloadAll() {
	err := postCommand(nil, rsTY.CommandUnloadAll)
	if err != nil {
		zap.L().Error("error on unload tasks command", zap.Error(err))
	}
}

// Enable task
func Enable(ids []string) error {
	tasks, err := getTaskEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(tasks); index++ {
		cfg := tasks[index]
		if !cfg.Enabled {
			cfg.Enabled = true
			err = SaveAndReload(&cfg)
			if err != nil {
				zap.L().Error("error on enabling a task", zap.String("id", cfg.ID), zap.Error(err))
			}
		}
	}
	return nil
}

// Disable task
func Disable(ids []string) error {
	tasks, err := getTaskEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(tasks); index++ {
		cfg := tasks[index]
		if cfg.Enabled {
			cfg.Enabled = false
			err = Save(&cfg)
			if err != nil {
				zap.L().Error("error on saving a task", zap.String("id", cfg.ID), zap.Error(err))
			}
			err = Remove(&cfg)
			if err != nil {
				zap.L().Error("error on disabling a task", zap.String("id", cfg.ID), zap.Error(err))
			}
		}
	}
	return nil
}

// Reload task
func Reload(ids []string) error {
	tasks, err := getTaskEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(tasks); index++ {
		task := tasks[index]
		err = Remove(&task)
		if err != nil {
			zap.L().Error("error on disabling a task", zap.Error(err), zap.String("id", task.ID))
		}
		if task.Enabled {
			err = Add(&task)
			if err != nil {
				zap.L().Error("error on enabling a task", zap.Error(err), zap.String("ic", task.ID))
			}
		}
	}
	return nil
}

func postCommand(cfg *taskTY.Config, command string) error {
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeTask,
		Command: command,
	}
	if cfg != nil {
		reqEvent.ID = cfg.ID
		reqEvent.SetData(cfg)
	}
	topic := mcbus.FormatTopic(mcbus.TopicServiceTask)
	return mcbus.Publish(topic, reqEvent)
}

func getTaskEntries(ids []string) ([]taskTY.Config, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	result, err := List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *result.Data.(*[]taskTY.Config), nil
}

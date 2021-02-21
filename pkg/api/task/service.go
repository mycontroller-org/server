package task

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

// Add task
func Add(cfg *taskML.Config) error {
	return postCommand(cfg, rsML.CommandAdd)
}

// Remove task
func Remove(cfg *taskML.Config) error {
	return postCommand(cfg, rsML.CommandRemove)
}

// LoadAll makes tasks alive
func LoadAll() {
	result, err := List(nil, nil)
	if err != nil {
		zap.L().Error("Failed to get list of tasks", zap.Error(err))
		return
	}
	tasks := *result.Data.(*[]taskML.Config)
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
	err := postCommand(nil, rsML.CommandUnloadAll)
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
			err = Save(&cfg)
			if err != nil {
				return err
			}
			return postCommand(&cfg, rsML.CommandAdd)
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

// Reload task
func Reload(ids []string) error {
	tasks, err := getTaskEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(tasks); index++ {
		task := tasks[index]
		err = postCommand(&task, rsML.CommandRemove)
		if err != nil {
			zap.L().Error("error on posting task remove command", zap.Error(err), zap.String("task", task.ID))
		}
		if task.Enabled {
			err = postCommand(&task, rsML.CommandAdd)
			if err != nil {
				zap.L().Error("error on posting task reload command", zap.Error(err), zap.String("task", task.ID))
			}
		}
	}
	return nil
}

func postCommand(cfg *taskML.Config, command string) error {
	reqEvent := rsML.Event{
		Type:    rsML.TypeTask,
		Command: command,
	}
	if cfg != nil {
		reqEvent.ID = cfg.ID
		err := reqEvent.SetData(cfg)
		if err != nil {
			return err
		}
	}
	topic := mcbus.FormatTopic(mcbus.TopicServiceTask)
	return mcbus.Publish(topic, reqEvent)
}

func getTaskEntries(ids []string) ([]taskML.Config, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: ids}}
	pagination := &stgml.Pagination{Limit: 100}
	result, err := List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *result.Data.(*[]taskML.Config), nil
}

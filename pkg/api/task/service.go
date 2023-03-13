package task

import (
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// Add task
func (t *TaskAPI) Add(task *taskTY.Config) error {
	return t.postCommand(task, rsTY.CommandAdd)
}

// Remove task
func (t *TaskAPI) Remove(task *taskTY.Config) error {
	return t.postCommand(task, rsTY.CommandRemove)
}

// LoadAll makes tasks alive
func (t *TaskAPI) LoadAll() {
	result, err := t.List(nil, nil)
	if err != nil {
		t.logger.Error("failed to get list of tasks", zap.Error(err))
		return
	}
	tasks := *result.Data.(*[]taskTY.Config)
	for index := 0; index < len(tasks); index++ {
		task := tasks[index]
		if task.Enabled || task.ReEnable {
			err = t.Add(&task)
			if err != nil {
				t.logger.Error("failed to load a task", zap.Error(err), zap.String("task", task.ID))
			}
		}
	}
}

// UnloadAll makes stop all tasks
func (t *TaskAPI) UnloadAll() {
	err := t.postCommand(nil, rsTY.CommandUnloadAll)
	if err != nil {
		t.logger.Error("error on unload tasks command", zap.Error(err))
	}
}

// Enable task
func (t *TaskAPI) Enable(ids []string) error {
	tasks, err := t.getTaskEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(tasks); index++ {
		task := tasks[index]
		task.Enabled = true
		err = t.SaveAndReload(&task)
		if err != nil {
			t.logger.Error("error on enabling a task", zap.String("id", task.ID), zap.Error(err))
		}
	}
	return nil
}

// Disable task
func (t *TaskAPI) Disable(ids []string) error {
	tasks, err := t.getTaskEntries(ids)
	if err != nil {
		return err
	}

	for index := 0; index < len(tasks); index++ {
		task := tasks[index]
		if task.Enabled {
			task.Enabled = false
			err = t.Save(&task)
			if err != nil {
				t.logger.Error("error on saving a task", zap.String("id", task.ID), zap.Error(err))
			}
			err = t.Remove(&task)
			if err != nil {
				t.logger.Error("error on disabling a task", zap.String("id", task.ID), zap.Error(err))
			}
			// if it is re-enable task load it
			if task.ReEnable {
				err = t.Add(&task)
				if err != nil {
					t.logger.Error("error on adding a re-enable task", zap.String("id", task.ID), zap.Error(err))
				}
			}
		}
	}
	return nil
}

// Reload task
func (t *TaskAPI) Reload(ids []string) error {
	tasks, err := t.getTaskEntries(ids)
	if err != nil {
		return err
	}
	for index := 0; index < len(tasks); index++ {
		task := tasks[index]
		err = t.Remove(&task)
		if err != nil {
			t.logger.Error("error on disabling a task", zap.Error(err), zap.String("id", task.ID))
		}
		if task.Enabled || task.ReEnable {
			err = t.Add(&task)
			if err != nil {
				t.logger.Error("error on enabling a task", zap.Error(err), zap.String("ic", task.ID))
			}
		}
	}
	return nil
}

func (t *TaskAPI) postCommand(task *taskTY.Config, command string) error {
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeTask,
		Command: command,
	}
	if task != nil {
		reqEvent.ID = task.ID
		reqEvent.SetData(task)
	}
	return t.bus.Publish(topic.TopicServiceTask, reqEvent)
}

func (t *TaskAPI) getTaskEntries(ids []string) ([]taskTY.Config, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	result, err := t.List(filters, pagination)
	if err != nil {
		return nil, err
	}
	return *result.Data.(*[]taskTY.Config), nil
}

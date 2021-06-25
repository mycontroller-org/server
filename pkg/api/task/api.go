package task

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
	taskML "github.com/mycontroller-org/server/v2/pkg/model/task"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	stg "github.com/mycontroller-org/server/v2/pkg/service/storage"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	stgML "github.com/mycontroller-org/server/v2/plugin/database/storage"
)

// List by filter and pagination
func List(filters []stgML.Filter, pagination *stgML.Pagination) (*stgML.Result, error) {
	result := make([]taskML.Config, 0)
	return stg.SVC.Find(model.EntityTask, &result, filters, pagination)
}

// Get returns a task
func Get(filters []stgML.Filter) (*taskML.Config, error) {
	result := &taskML.Config{}
	err := stg.SVC.FindOne(model.EntityTask, result, filters)
	return result, err
}

// Save a task details
func Save(task *taskML.Config) error {
	eventType := eventML.TypeUpdated
	if task.ID == "" {
		task.ID = utils.RandUUID()
		eventType = eventML.TypeCreated
	}
	filters := []stgML.Filter{
		{Key: model.KeyID, Value: task.ID},
	}
	err := stg.SVC.Upsert(model.EntityTask, task, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventTask, eventType, model.EntityTask, task)
	return nil
}

// SaveAndReload task
func SaveAndReload(cfg *taskML.Config) error {
	cfg.State = &taskML.State{} // reset state
	err := Save(cfg)
	if err != nil {
		return err
	}
	return Reload([]string{cfg.ID})
}

// GetByID returns a task by id
func GetByID(id string) (*taskML.Config, error) {
	f := []stgML.Filter{
		{Key: model.KeyID, Value: id},
	}
	out := &taskML.Config{}
	err := stg.SVC.FindOne(model.EntityTask, out, f)
	return out, err
}

// SetState Updates state data
func SetState(id string, state *taskML.State) error {
	task, err := GetByID(id)
	if err != nil {
		return err
	}
	task.State = state
	return Save(task)
}

// Delete tasks
func Delete(IDs []string) (int64, error) {
	err := Disable(IDs)
	if err != nil {
		return 0, err
	}
	filters := []stgML.Filter{{Key: model.KeyID, Operator: stgML.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(model.EntityTask, filters)
}

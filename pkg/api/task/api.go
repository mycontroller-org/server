package task

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
	taskML "github.com/mycontroller-org/server/v2/pkg/model/task"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// List by filter and pagination
func List(filters []stgType.Filter, pagination *stgType.Pagination) (*stgType.Result, error) {
	result := make([]taskML.Config, 0)
	return store.STORAGE.Find(model.EntityTask, &result, filters, pagination)
}

// Get returns a task
func Get(filters []stgType.Filter) (*taskML.Config, error) {
	result := &taskML.Config{}
	err := store.STORAGE.FindOne(model.EntityTask, result, filters)
	return result, err
}

// Save a task details
func Save(task *taskML.Config) error {
	eventType := eventML.TypeUpdated
	if task.ID == "" {
		task.ID = utils.RandUUID()
		eventType = eventML.TypeCreated
	}
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: task.ID},
	}
	err := store.STORAGE.Upsert(model.EntityTask, task, filters)
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
	f := []stgType.Filter{
		{Key: model.KeyID, Value: id},
	}
	out := &taskML.Config{}
	err := store.STORAGE.FindOne(model.EntityTask, out, f)
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
	filters := []stgType.Filter{{Key: model.KeyID, Operator: stgType.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(model.EntityTask, filters)
}

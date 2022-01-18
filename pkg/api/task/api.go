package task

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]taskTY.Config, 0)
	return store.STORAGE.Find(types.EntityTask, &result, filters, pagination)
}

// Get returns a task
func Get(filters []storageTY.Filter) (*taskTY.Config, error) {
	result := &taskTY.Config{}
	err := store.STORAGE.FindOne(types.EntityTask, result, filters)
	return result, err
}

// Save a task details
func Save(task *taskTY.Config) error {
	eventType := eventTY.TypeUpdated
	if task.ID == "" {
		task.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: task.ID},
	}
	err := store.STORAGE.Upsert(types.EntityTask, task, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventTask, eventType, types.EntityTask, task)
	return nil
}

// SaveAndReload task
func SaveAndReload(cfg *taskTY.Config) error {
	cfg.State = &taskTY.State{} // reset state
	err := Save(cfg)
	if err != nil {
		return err
	}
	return Reload([]string{cfg.ID})
}

// GetByID returns a task by id
func GetByID(id string) (*taskTY.Config, error) {
	f := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	out := &taskTY.Config{}
	err := store.STORAGE.FindOne(types.EntityTask, out, f)
	return out, err
}

// SetState Updates state data
func SetState(id string, state *taskTY.State) error {
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
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(types.EntityTask, filters)
}

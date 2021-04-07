package task

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]taskML.Config, 0)
	return stg.SVC.Find(ml.EntityTask, &result, filters, pagination)
}

// Get returns a task
func Get(filters []stgml.Filter) (*taskML.Config, error) {
	result := &taskML.Config{}
	err := stg.SVC.FindOne(ml.EntityTask, result, filters)
	return result, err
}

// Save a task details
func Save(task *taskML.Config) error {
	if task.ID == "" {
		task.ID = ut.RandUUID()
	}
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: task.ID},
	}
	err := stg.SVC.Upsert(ml.EntityTask, task, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventTask, *task)
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
	f := []stgml.Filter{
		{Key: ml.KeyID, Value: id},
	}
	out := &taskML.Config{}
	err := stg.SVC.FindOne(ml.EntityTask, out, f)
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
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(ml.EntityTask, filters)
}

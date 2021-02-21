package scheduler

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	schedulerML "github.com/mycontroller-org/backend/v2/pkg/model/scheduler"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]schedulerML.Config, 0)
	return stg.SVC.Find(ml.EntityScheduler, &result, filters, pagination)
}

// Get returns a scheduler
func Get(filters []stgml.Filter) (*schedulerML.Config, error) {
	result := &schedulerML.Config{}
	err := stg.SVC.FindOne(ml.EntityScheduler, result, filters)
	return result, err
}

// Save a scheduler details
func Save(scheduler *schedulerML.Config) error {
	if scheduler.ID == "" {
		scheduler.ID = ut.RandUUID()
	}
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: scheduler.ID},
	}
	return stg.SVC.Upsert(ml.EntityScheduler, scheduler, filters)
}

// SaveAndReload scheduler
func SaveAndReload(cfg *schedulerML.Config) error {
	cfg.State = &schedulerML.State{} // reset state
	err := Save(cfg)
	if err != nil {
		return err
	}
	return Reload([]string{cfg.ID})
}

// GetByID returns a scheduler by id
func GetByID(id string) (*schedulerML.Config, error) {
	f := []stgml.Filter{
		{Key: ml.KeyID, Value: id},
	}
	out := &schedulerML.Config{}
	err := stg.SVC.FindOne(ml.EntityScheduler, out, f)
	return out, err
}

// SetState Updates state data
func SetState(id string, state *schedulerML.State) error {
	scheduler, err := GetByID(id)
	if err != nil {
		return err
	}
	scheduler.State = state
	return Save(scheduler)
}

// Delete schedulers
func Delete(IDs []string) (int64, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(ml.EntityScheduler, filters)
}

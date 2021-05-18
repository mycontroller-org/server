package scheduler

import (
	"github.com/mycontroller-org/backend/v2/pkg/model"
	eventML "github.com/mycontroller-org/backend/v2/pkg/model/bus/event"
	schedulerML "github.com/mycontroller-org/backend/v2/pkg/model/scheduler"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	stgML "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgML.Filter, pagination *stgML.Pagination) (*stgML.Result, error) {
	result := make([]schedulerML.Config, 0)
	return stg.SVC.Find(model.EntityScheduler, &result, filters, pagination)
}

// Get returns a scheduler
func Get(filters []stgML.Filter) (*schedulerML.Config, error) {
	result := &schedulerML.Config{}
	err := stg.SVC.FindOne(model.EntityScheduler, result, filters)
	return result, err
}

// Save a scheduler details
func Save(schedule *schedulerML.Config) error {
	eventType := eventML.TypeUpdated
	if schedule.ID == "" {
		schedule.ID = utils.RandUUID()
		eventType = eventML.TypeCreated
	}

	filters := []stgML.Filter{
		{Key: model.KeyID, Value: schedule.ID},
	}
	err := stg.SVC.Upsert(model.EntityScheduler, schedule, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventSchedule, eventType, model.EntityHandler, *schedule)
	return nil
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
	filters := []stgML.Filter{
		{Key: model.KeyID, Value: id},
	}
	out := &schedulerML.Config{}
	err := stg.SVC.FindOne(model.EntityScheduler, out, filters)
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
	// disable the schedules
	err := Disable(IDs)
	if err != nil {
		return 0, err
	}
	filters := []stgML.Filter{{Key: model.KeyID, Operator: stgML.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(model.EntityScheduler, filters)
}

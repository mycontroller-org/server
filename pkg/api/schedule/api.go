package schedule

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
	scheduleML "github.com/mycontroller-org/server/v2/pkg/model/schedule"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	stg "github.com/mycontroller-org/server/v2/pkg/service/storage"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	stgML "github.com/mycontroller-org/server/v2/plugin/database/storage"
)

// List by filter and pagination
func List(filters []stgML.Filter, pagination *stgML.Pagination) (*stgML.Result, error) {
	result := make([]scheduleML.Config, 0)
	return stg.SVC.Find(model.EntitySchedule, &result, filters, pagination)
}

// Get returns a scheduler
func Get(filters []stgML.Filter) (*scheduleML.Config, error) {
	result := &scheduleML.Config{}
	err := stg.SVC.FindOne(model.EntitySchedule, result, filters)
	return result, err
}

// Save a scheduler details
func Save(schedule *scheduleML.Config) error {
	eventType := eventML.TypeUpdated
	if schedule.ID == "" {
		schedule.ID = utils.RandUUID()
		eventType = eventML.TypeCreated
	}

	filters := []stgML.Filter{
		{Key: model.KeyID, Value: schedule.ID},
	}
	err := stg.SVC.Upsert(model.EntitySchedule, schedule, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventSchedule, eventType, model.EntityHandler, *schedule)
	return nil
}

// SaveAndReload scheduler
func SaveAndReload(cfg *scheduleML.Config) error {
	cfg.State = &scheduleML.State{} // reset state
	err := Save(cfg)
	if err != nil {
		return err
	}
	return Reload([]string{cfg.ID})
}

// GetByID returns a scheduler by id
func GetByID(id string) (*scheduleML.Config, error) {
	filters := []stgML.Filter{
		{Key: model.KeyID, Value: id},
	}
	out := &scheduleML.Config{}
	err := stg.SVC.FindOne(model.EntitySchedule, out, filters)
	return out, err
}

// SetState Updates state data
func SetState(id string, state *scheduleML.State) error {
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
	return stg.SVC.Delete(model.EntitySchedule, filters)
}

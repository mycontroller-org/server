package schedule

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
	scheduleML "github.com/mycontroller-org/server/v2/pkg/model/schedule"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// List by filter and pagination
func List(filters []stgType.Filter, pagination *stgType.Pagination) (*stgType.Result, error) {
	result := make([]scheduleML.Config, 0)
	return store.STORAGE.Find(model.EntitySchedule, &result, filters, pagination)
}

// Get returns a scheduler
func Get(filters []stgType.Filter) (*scheduleML.Config, error) {
	result := &scheduleML.Config{}
	err := store.STORAGE.FindOne(model.EntitySchedule, result, filters)
	return result, err
}

// Save a scheduler details
func Save(schedule *scheduleML.Config) error {
	eventType := eventML.TypeUpdated
	if schedule.ID == "" {
		schedule.ID = utils.RandUUID()
		eventType = eventML.TypeCreated
	}

	filters := []stgType.Filter{
		{Key: model.KeyID, Value: schedule.ID},
	}
	err := store.STORAGE.Upsert(model.EntitySchedule, schedule, filters)
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
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: id},
	}
	out := &scheduleML.Config{}
	err := store.STORAGE.FindOne(model.EntitySchedule, out, filters)
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
	filters := []stgType.Filter{{Key: model.KeyID, Operator: stgType.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(model.EntitySchedule, filters)
}

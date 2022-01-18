package schedule

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]scheduleTY.Config, 0)
	return store.STORAGE.Find(types.EntitySchedule, &result, filters, pagination)
}

// Get returns a scheduler
func Get(filters []storageTY.Filter) (*scheduleTY.Config, error) {
	result := &scheduleTY.Config{}
	err := store.STORAGE.FindOne(types.EntitySchedule, result, filters)
	return result, err
}

// Save a scheduler details
func Save(schedule *scheduleTY.Config) error {
	eventType := eventTY.TypeUpdated
	if schedule.ID == "" {
		schedule.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: schedule.ID},
	}
	err := store.STORAGE.Upsert(types.EntitySchedule, schedule, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventSchedule, eventType, types.EntityHandler, *schedule)
	return nil
}

// SaveAndReload scheduler
func SaveAndReload(cfg *scheduleTY.Config) error {
	cfg.State = &scheduleTY.State{} // reset state
	err := Save(cfg)
	if err != nil {
		return err
	}
	return Reload([]string{cfg.ID})
}

// GetByID returns a scheduler by id
func GetByID(id string) (*scheduleTY.Config, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	out := &scheduleTY.Config{}
	err := store.STORAGE.FindOne(types.EntitySchedule, out, filters)
	return out, err
}

// SetState Updates state data
func SetState(id string, state *scheduleTY.State) error {
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
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(types.EntitySchedule, filters)
}

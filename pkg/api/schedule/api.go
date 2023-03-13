package schedule

import (
	"context"
	"errors"
	"fmt"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type ScheduleAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	bus     busTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, bus busTY.Plugin) *ScheduleAPI {
	return &ScheduleAPI{
		ctx:     ctx,
		logger:  logger.Named("schedule_api"),
		storage: storage,
		bus:     bus,
	}
}

// List by filter and pagination
func (sh *ScheduleAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]schedulerTY.Config, 0)
	return sh.storage.Find(types.EntitySchedule, &result, filters, pagination)
}

// Get returns a scheduler
func (sh *ScheduleAPI) Get(filters []storageTY.Filter) (*schedulerTY.Config, error) {
	result := &schedulerTY.Config{}
	err := sh.storage.FindOne(types.EntitySchedule, result, filters)
	return result, err
}

// Save a scheduler details
func (sh *ScheduleAPI) Save(schedule *schedulerTY.Config) error {
	eventType := eventTY.TypeUpdated
	if schedule.ID == "" {
		schedule.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: schedule.ID},
	}
	err := sh.storage.Upsert(types.EntitySchedule, schedule, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(sh.logger, sh.bus, topic.TopicEventSchedule, eventType, types.EntitySchedule, *schedule)
	return nil
}

// SaveAndReload scheduler
func (sh *ScheduleAPI) SaveAndReload(cfg *schedulerTY.Config) error {
	cfg.State = &schedulerTY.State{} // reset state
	err := sh.Save(cfg)
	if err != nil {
		return err
	}
	return sh.Reload([]string{cfg.ID})
}

// GetByID returns a scheduler by id
func (sh *ScheduleAPI) GetByID(id string) (*schedulerTY.Config, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	out := &schedulerTY.Config{}
	err := sh.storage.FindOne(types.EntitySchedule, out, filters)
	return out, err
}

// SetState Updates state data
func (sh *ScheduleAPI) SetState(id string, state *schedulerTY.State) error {
	scheduler, err := sh.GetByID(id)
	if err != nil {
		return err
	}
	scheduler.State = state
	return sh.Save(scheduler)
}

// Delete schedulers
func (sh *ScheduleAPI) Delete(IDs []string) (int64, error) {
	// disable the schedules
	err := sh.Disable(IDs)
	if err != nil {
		return 0, err
	}
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return sh.storage.Delete(types.EntitySchedule, filters)
}

func (sh *ScheduleAPI) Import(data interface{}) error {
	input, ok := data.(schedulerTY.Config)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		return errors.New("'id' can not be empty")
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	return sh.storage.Upsert(types.EntitySchedule, &input, filters)
}

func (sh *ScheduleAPI) GetEntityInterface() interface{} {
	return schedulerTY.Config{}
}

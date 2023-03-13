package task

import (
	"context"
	"errors"
	"fmt"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type TaskAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	bus     busTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, bus busTY.Plugin) *TaskAPI {
	return &TaskAPI{
		ctx:     ctx,
		logger:  logger.Named("task_api"),
		storage: storage,
		bus:     bus,
	}
}

// List by filter and pagination
func (t *TaskAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]taskTY.Config, 0)
	return t.storage.Find(types.EntityTask, &result, filters, pagination)
}

// Get returns a task
func (t *TaskAPI) Get(filters []storageTY.Filter) (*taskTY.Config, error) {
	result := &taskTY.Config{}
	err := t.storage.FindOne(types.EntityTask, result, filters)
	return result, err
}

// Save a task details
func (t *TaskAPI) Save(task *taskTY.Config) error {
	eventType := eventTY.TypeUpdated
	if task.ID == "" {
		task.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: task.ID},
	}
	err := t.storage.Upsert(types.EntityTask, task, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(t.logger, t.bus, topic.TopicEventTask, eventType, types.EntityTask, task)
	return nil
}

// SaveAndReload task
func (t *TaskAPI) SaveAndReload(cfg *taskTY.Config) error {
	cfg.State = &taskTY.State{} // reset state
	err := t.Save(cfg)
	if err != nil {
		return err
	}
	return t.Reload([]string{cfg.ID})
}

// GetByID returns a task by id
func (t *TaskAPI) GetByID(id string) (*taskTY.Config, error) {
	f := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	out := &taskTY.Config{}
	err := t.storage.FindOne(types.EntityTask, out, f)
	return out, err
}

// SetState Updates state data
func (t *TaskAPI) SetState(id string, state *taskTY.State) error {
	task, err := t.GetByID(id)
	if err != nil {
		return err
	}
	task.State = state
	return t.Save(task)
}

// Delete tasks
func (t *TaskAPI) Delete(IDs []string) (int64, error) {
	err := t.Disable(IDs)
	if err != nil {
		return 0, err
	}
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return t.storage.Delete(types.EntityTask, filters)
}

func (t *TaskAPI) Import(data interface{}) error {
	input, ok := data.(taskTY.Config)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		return errors.New("'id' can not be empty")
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	return t.storage.Upsert(types.EntityTask, &input, filters)
}

func (t *TaskAPI) GetEntityInterface() interface{} {
	return taskTY.Config{}
}

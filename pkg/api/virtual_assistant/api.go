package assistant

import (
	"context"
	"errors"
	"fmt"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	vaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type VirtualAssistantAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	bus     busTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, bus busTY.Plugin) *VirtualAssistantAPI {
	return &VirtualAssistantAPI{
		ctx:     ctx,
		logger:  logger.Named("virtual_assistant_api"),
		storage: storage,
		bus:     bus,
	}
}

// List by filter and pagination
func (va *VirtualAssistantAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]vaTY.Config, 0)
	return va.storage.Find(types.EntityVirtualAssistant, &result, filters, pagination)
}

// Get returns a virtual assistant
func (va *VirtualAssistantAPI) Get(filters []storageTY.Filter) (*vaTY.Config, error) {
	result := &vaTY.Config{}
	err := va.storage.FindOne(types.EntityVirtualAssistant, result, filters)
	return result, err
}

// Save a virtual assistant details
func (va *VirtualAssistantAPI) Save(cfg *vaTY.Config) error {
	eventType := eventTY.TypeUpdated
	if cfg.ID == "" {
		cfg.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: cfg.ID},
	}

	cfg.ModifiedOn = time.Now()

	err := va.storage.Upsert(types.EntityVirtualAssistant, cfg, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(va.logger, va.bus, topic.TopicEventVirtualAssistant, eventType, types.EntityVirtualAssistant, *cfg)
	return nil
}

// SaveAndReload virtual assistant
func (va *VirtualAssistantAPI) SaveAndReload(cfg *vaTY.Config) error {
	cfg.State = &types.State{} // reset state
	err := va.Save(cfg)
	if err != nil {
		return err
	}
	return va.Reload([]string{cfg.ID})
}

// GetByID returns a virtual assistant by id
func (va *VirtualAssistantAPI) GetByID(id string) (*vaTY.Config, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	out := &vaTY.Config{}
	err := va.storage.FindOne(types.EntityVirtualAssistant, out, filters)
	return out, err
}

// SetState Updates state data
func (va *VirtualAssistantAPI) SetState(id string, state *types.State) error {
	cfg, err := va.GetByID(id)
	if err != nil {
		return err
	}
	cfg.State = state
	return va.Save(cfg)
}

// Delete virtual assistants
func (va *VirtualAssistantAPI) Delete(IDs []string) (int64, error) {
	// disable virtual assistants
	err := va.Disable(IDs)
	if err != nil {
		return 0, err
	}
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return va.storage.Delete(types.EntityVirtualAssistant, filters)
}

func (va *VirtualAssistantAPI) Import(data interface{}) error {
	input, ok := data.(vaTY.Config)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		return errors.New("'id' can not be empty")
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	return va.storage.Upsert(types.EntityVirtualAssistant, &input, filters)
}

func (va *VirtualAssistantAPI) GetEntityInterface() interface{} {
	return vaTY.Config{}
}

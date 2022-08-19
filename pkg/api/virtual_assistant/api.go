package assistant

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	vaTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_assistant"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]vaTY.Config, 0)
	return store.STORAGE.Find(types.EntityVirtualAssistant, &result, filters, pagination)
}

// Get returns a virtual assistant
func Get(filters []storageTY.Filter) (*vaTY.Config, error) {
	result := &vaTY.Config{}
	err := store.STORAGE.FindOne(types.EntityVirtualAssistant, result, filters)
	return result, err
}

// Save a virtual assistant details
func Save(cfg *vaTY.Config) error {
	eventType := eventTY.TypeUpdated
	if cfg.ID == "" {
		cfg.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: cfg.ID},
	}

	cfg.ModifiedOn = time.Now()

	err := store.STORAGE.Upsert(types.EntityVirtualAssistant, cfg, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventVirtualAssistant, eventType, types.EntityVirtualAssistant, *cfg)
	return nil
}

// SaveAndReload virtual assistant
func SaveAndReload(cfg *vaTY.Config) error {
	cfg.State = &types.State{} // reset state
	err := Save(cfg)
	if err != nil {
		return err
	}
	return Reload([]string{cfg.ID})
}

// GetByID returns a virtual assistant by id
func GetByID(id string) (*vaTY.Config, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	out := &vaTY.Config{}
	err := store.STORAGE.FindOne(types.EntityVirtualAssistant, out, filters)
	return out, err
}

// SetState Updates state data
func SetState(id string, state *types.State) error {
	cfg, err := GetByID(id)
	if err != nil {
		return err
	}
	cfg.State = state
	return Save(cfg)
}

// Delete virtual assistants
func Delete(IDs []string) (int64, error) {
	// disable virtual assistants
	err := Disable(IDs)
	if err != nil {
		return 0, err
	}
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(types.EntityVirtualAssistant, filters)
}

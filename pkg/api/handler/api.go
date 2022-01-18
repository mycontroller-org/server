package handler

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	out := make([]handlerTY.Config, 0)
	return store.STORAGE.Find(types.EntityHandler, &out, filters, pagination)
}

// Get a config
func Get(f []storageTY.Filter) (handlerTY.Config, error) {
	out := handlerTY.Config{}
	err := store.STORAGE.FindOne(types.EntityHandler, &out, f)
	return out, err
}

// SaveAndReload handler
func SaveAndReload(cfg *handlerTY.Config) error {
	cfg.State = &types.State{} // reset state
	err := Save(cfg)
	if err != nil {
		return err
	}
	return Reload([]string{cfg.ID})
}

// Save config
func Save(cfg *handlerTY.Config) error {
	eventType := eventTY.TypeUpdated
	if cfg.ID == "" {
		cfg.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}

	// encrypt passwords
	err := cloneUtil.UpdateSecrets(cfg, store.CFG.Secret, "", true, cloneUtil.DefaultSpecialKeys)
	if err != nil {
		return err
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: cfg.ID},
	}
	err = store.STORAGE.Upsert(types.EntityHandler, cfg, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventHandler, eventType, types.EntityHandler, cfg)
	return nil
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

// GetByTypeName returns a handler by type and name
func GetByTypeName(handlerPluginType, name string) (*handlerTY.Config, error) {
	f := []storageTY.Filter{
		{Key: types.KeyHandlerType, Value: handlerPluginType},
		{Key: types.KeyHandlerName, Value: name},
	}
	out := &handlerTY.Config{}
	err := store.STORAGE.FindOne(types.EntityHandler, out, f)
	return out, err
}

// GetByID returns a handler by id
func GetByID(ID string) (*handlerTY.Config, error) {
	f := []storageTY.Filter{
		{Key: types.KeyID, Value: ID},
	}
	out := &handlerTY.Config{}
	err := store.STORAGE.FindOne(types.EntityHandler, out, f)
	return out, err
}

// Delete Service
func Delete(ids []string) (int64, error) {
	err := Disable(ids)
	if err != nil {
		return 0, err
	}
	f := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	return store.STORAGE.Delete(types.EntityHandler, f)
}

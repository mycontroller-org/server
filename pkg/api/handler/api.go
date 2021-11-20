package handler

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/type"
)

// List by filter and pagination
func List(filters []stgType.Filter, pagination *stgType.Pagination) (*stgType.Result, error) {
	out := make([]handlerType.Config, 0)
	return store.STORAGE.Find(model.EntityHandler, &out, filters, pagination)
}

// Get a config
func Get(f []stgType.Filter) (handlerType.Config, error) {
	out := handlerType.Config{}
	err := store.STORAGE.FindOne(model.EntityHandler, &out, f)
	return out, err
}

// SaveAndReload handler
func SaveAndReload(cfg *handlerType.Config) error {
	cfg.State = &model.State{} // reset state
	err := Save(cfg)
	if err != nil {
		return err
	}
	return Reload([]string{cfg.ID})
}

// Save config
func Save(cfg *handlerType.Config) error {
	eventType := eventML.TypeUpdated
	if cfg.ID == "" {
		cfg.ID = utils.RandUUID()
		eventType = eventML.TypeCreated
	}

	// encrypt passwords
	err := cloneUtil.UpdateSecrets(cfg, store.CFG.Secret, true)
	if err != nil {
		return err
	}

	filters := []stgType.Filter{
		{Key: model.KeyID, Value: cfg.ID},
	}
	err = store.STORAGE.Upsert(model.EntityHandler, cfg, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventHandler, eventType, model.EntityHandler, cfg)
	return nil
}

// SetState Updates state data
func SetState(id string, state *model.State) error {
	cfg, err := GetByID(id)
	if err != nil {
		return err
	}
	cfg.State = state
	return Save(cfg)
}

// GetByTypeName returns a handler by type and name
func GetByTypeName(handlerPluginType, name string) (*handlerType.Config, error) {
	f := []stgType.Filter{
		{Key: model.KeyHandlerType, Value: handlerPluginType},
		{Key: model.KeyHandlerName, Value: name},
	}
	out := &handlerType.Config{}
	err := store.STORAGE.FindOne(model.EntityHandler, out, f)
	return out, err
}

// GetByID returns a handler by id
func GetByID(ID string) (*handlerType.Config, error) {
	f := []stgType.Filter{
		{Key: model.KeyID, Value: ID},
	}
	out := &handlerType.Config{}
	err := store.STORAGE.FindOne(model.EntityHandler, out, f)
	return out, err
}

// Delete Service
func Delete(ids []string) (int64, error) {
	err := Disable(ids)
	if err != nil {
		return 0, err
	}
	f := []stgType.Filter{{Key: model.KeyID, Operator: stgType.OperatorIn, Value: ids}}
	return store.STORAGE.Delete(model.EntityHandler, f)
}

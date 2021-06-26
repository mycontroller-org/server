package handler

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
	handlerML "github.com/mycontroller-org/server/v2/pkg/model/handler"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	stg "github.com/mycontroller-org/server/v2/pkg/service/database/storage"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	stgML "github.com/mycontroller-org/server/v2/plugin/database/storage"
)

// List by filter and pagination
func List(filters []stgML.Filter, pagination *stgML.Pagination) (*stgML.Result, error) {
	out := make([]handlerML.Config, 0)
	return stg.SVC.Find(model.EntityHandler, &out, filters, pagination)
}

// Get a config
func Get(f []stgML.Filter) (handlerML.Config, error) {
	out := handlerML.Config{}
	err := stg.SVC.FindOne(model.EntityHandler, &out, f)
	return out, err
}

// SaveAndReload handler
func SaveAndReload(cfg *handlerML.Config) error {
	cfg.State = &model.State{} // reset state
	err := Save(cfg)
	if err != nil {
		return err
	}
	return Reload([]string{cfg.ID})
}

// Save config
func Save(cfg *handlerML.Config) error {
	eventType := eventML.TypeUpdated
	if cfg.ID == "" {
		cfg.ID = utils.RandUUID()
		eventType = eventML.TypeCreated
	}

	// encrypt passwords
	err := cloneUtil.UpdateSecrets(cfg, true)
	if err != nil {
		return err
	}

	filters := []stgML.Filter{
		{Key: model.KeyID, Value: cfg.ID},
	}
	err = stg.SVC.Upsert(model.EntityHandler, cfg, filters)
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
func GetByTypeName(handlerType, name string) (*handlerML.Config, error) {
	f := []stgML.Filter{
		{Key: model.KeyHandlerType, Value: handlerType},
		{Key: model.KeyHandlerName, Value: name},
	}
	out := &handlerML.Config{}
	err := stg.SVC.FindOne(model.EntityHandler, out, f)
	return out, err
}

// GetByID returns a handler by id
func GetByID(ID string) (*handlerML.Config, error) {
	f := []stgML.Filter{
		{Key: model.KeyID, Value: ID},
	}
	out := &handlerML.Config{}
	err := stg.SVC.FindOne(model.EntityHandler, out, f)
	return out, err
}

// Delete Service
func Delete(ids []string) (int64, error) {
	err := Disable(ids)
	if err != nil {
		return 0, err
	}
	f := []stgML.Filter{{Key: model.KeyID, Operator: stgML.OperatorIn, Value: ids}}
	return stg.SVC.Delete(model.EntityHandler, f)
}

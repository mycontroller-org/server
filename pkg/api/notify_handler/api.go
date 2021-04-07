package handler

import (
	"github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	out := make([]handlerML.Config, 0)
	return stg.SVC.Find(model.EntityNotifyHandler, &out, filters, pagination)
}

// Get a config
func Get(f []stgml.Filter) (handlerML.Config, error) {
	out := handlerML.Config{}
	err := stg.SVC.FindOne(model.EntityNotifyHandler, &out, f)
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
	if cfg.ID == "" {
		cfg.ID = ut.RandUUID()
	}
	f := []stgml.Filter{
		{Key: model.KeyID, Value: cfg.ID},
	}
	err := stg.SVC.Upsert(model.EntityNotifyHandler, cfg, f)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventHandler, *cfg)
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
	f := []stgml.Filter{
		{Key: model.KeyHandlerType, Value: handlerType},
		{Key: model.KeyHandlerName, Value: name},
	}
	out := &handlerML.Config{}
	err := stg.SVC.FindOne(model.EntityNotifyHandler, out, f)
	return out, err
}

// GetByID returns a handler by id
func GetByID(ID string) (*handlerML.Config, error) {
	f := []stgml.Filter{
		{Key: model.KeyID, Value: ID},
	}
	out := &handlerML.Config{}
	err := stg.SVC.FindOne(model.EntityNotifyHandler, out, f)
	return out, err
}

// Delete Service
func Delete(ids []string) (int64, error) {
	err := Disable(ids)
	if err != nil {
		return 0, err
	}
	f := []stgml.Filter{{Key: model.KeyID, Operator: stgml.OperatorIn, Value: ids}}
	return stg.SVC.Delete(model.EntityNotifyHandler, f)
}

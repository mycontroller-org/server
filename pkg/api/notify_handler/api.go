package handler

import (
	"github.com/mycontroller-org/backend/v2/pkg/model"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	out := make([]handlerML.Config, 0)
	return stg.SVC.Find(ml.EntityNotifyHandler, &out, filters, pagination)
}

// Get a config
func Get(f []stgml.Filter) (handlerML.Config, error) {
	out := handlerML.Config{}
	err := stg.SVC.FindOne(ml.EntityNotifyHandler, &out, f)
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
func Save(node *handlerML.Config) error {
	if node.ID == "" {
		node.ID = ut.RandUUID()
	}
	f := []stgml.Filter{
		{Key: ml.KeyID, Value: node.ID},
	}
	return stg.SVC.Upsert(ml.EntityNotifyHandler, node, f)
}

// SetState Updates state data
func SetState(id string, state *ml.State) error {
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
		{Key: ml.KeyHandlerType, Value: handlerType},
		{Key: ml.KeyHandlerName, Value: name},
	}
	out := &handlerML.Config{}
	err := stg.SVC.FindOne(ml.EntityNotifyHandler, out, f)
	return out, err
}

// GetByID returns a handler by id
func GetByID(ID string) (*handlerML.Config, error) {
	f := []stgml.Filter{
		{Key: ml.KeyID, Value: ID},
	}
	out := &handlerML.Config{}
	err := stg.SVC.FindOne(ml.EntityNotifyHandler, out, f)
	return out, err
}

// Delete Service
func Delete(ids []string) (int64, error) {
	err := Disable(ids)
	if err != nil {
		return 0, err
	}
	f := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: ids}}
	return stg.SVC.Delete(ml.EntityNotifyHandler, f)
}

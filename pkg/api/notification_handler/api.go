package services

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	nthlml "github.com/mycontroller-org/backend/v2/pkg/model/notification_handler"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(f []stgml.Filter, p *stgml.Pagination) (*stgml.Result, error) {
	out := make([]nthlml.Config, 0)
	return stg.SVC.Find(ml.EntityNotificationHandlers, &out, f, p)
}

// Get a Service
func Get(f []stgml.Filter) (nthlml.Config, error) {
	out := nthlml.Config{}
	err := stg.SVC.FindOne(ml.EntityNotificationHandlers, &out, f)
	return out, err
}

// Save Service config
func Save(node *nthlml.Config) error {
	if node.ID == "" {
		node.ID = ut.RandUUID()
	}
	f := []stgml.Filter{
		{Key: ml.KeyID, Value: node.ID},
	}
	return stg.SVC.Upsert(ml.EntityNotificationHandlers, node, f)
}

// GetByTypeName returns a Service by type and name
func GetByTypeName(ServiceType, name string) (*nthlml.Config, error) {
	f := []stgml.Filter{
		{Key: ml.KeyServiceType, Value: ServiceType},
		{Key: ml.KeyServiceName, Value: name},
	}
	out := &nthlml.Config{}
	err := stg.SVC.FindOne(ml.EntityNotificationHandlers, out, f)
	return out, err
}

// Delete Service
func Delete(IDs []string) (int64, error) {
	f := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(ml.EntityNotificationHandlers, f)
}

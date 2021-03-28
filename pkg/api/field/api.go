package field

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]fml.Field, 0)
	return stg.SVC.Find(ml.EntityField, &result, filters, pagination)
}

// Get returns a field
func Get(filters []stgml.Filter) (*fml.Field, error) {
	result := &fml.Field{}
	err := stg.SVC.FindOne(ml.EntityField, result, filters)
	return result, err
}

// Save a field details
func Save(field *fml.Field) error {
	if field.ID == "" {
		field.ID = ut.RandUUID()
	}
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: field.ID},
	}
	return stg.SVC.Upsert(ml.EntityField, field, filters)
}

// GetByIDs returns a field details by gatewayID, nodeId, sourceID and fieldName of a message
func GetByIDs(gatewayID, nodeID, sourceID, fieldID string) (*fml.Field, error) {
	filters := []stgml.Filter{
		{Key: ml.KeyGatewayID, Value: gatewayID},
		{Key: ml.KeyNodeID, Value: nodeID},
		{Key: ml.KeySourceID, Value: sourceID},
		{Key: ml.KeyFieldID, Value: fieldID},
	}
	result := &fml.Field{}
	err := stg.SVC.FindOne(ml.EntityField, result, filters)
	return result, err
}

// Delete fields
func Delete(IDs []string) (int64, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(ml.EntityField, filters)
}

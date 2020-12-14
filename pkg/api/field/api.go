package field

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]fml.Field, 0)
	return svc.STG.Find(ml.EntitySensorField, &result, filters, pagination)
}

// Get returns a field
func Get(filters []stgml.Filter) (*fml.Field, error) {
	result := &fml.Field{}
	err := svc.STG.FindOne(ml.EntitySensorField, result, filters)
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
	return svc.STG.Upsert(ml.EntitySensorField, field, filters)
}

// GetByIDs returns a field details by gatewayID, nodeId, sensorID and fieldName of a message
func GetByIDs(gatewayID, nodeID, sensorID, fieldID string) (*fml.Field, error) {
	filters := []stgml.Filter{
		{Key: ml.KeyGatewayID, Value: gatewayID},
		{Key: ml.KeyNodeID, Value: nodeID},
		{Key: ml.KeySensorID, Value: sensorID},
		{Key: ml.KeyFieldID, Value: fieldID},
	}
	result := &fml.Field{}
	err := svc.STG.FindOne(ml.EntitySensorField, result, filters)
	return result, err
}

// Delete sensor fields
func Delete(IDs []string) (int64, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return svc.STG.Delete(ml.EntitySensorField, filters)
}

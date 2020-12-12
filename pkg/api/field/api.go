package field

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	stgml "github.com/mycontroller-org/backend/v2/pkg/model/storage"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
)

// List by filter and pagination
func List(filters []pml.Filter, pagination *pml.Pagination) (*pml.Result, error) {
	result := make([]fml.Field, 0)
	return svc.STG.Find(ml.EntitySensorField, &result, filters, pagination)
}

// Get returns a field
func Get(filters []pml.Filter) (*fml.Field, error) {
	result := &fml.Field{}
	err := svc.STG.FindOne(ml.EntitySensorField, result, filters)
	return result, err
}

// Save a field details
func Save(field *fml.Field) error {
	if field.ID == "" {
		field.ID = ut.RandUUID()
	}
	filters := []pml.Filter{
		{Key: ml.KeyID, Value: field.ID},
	}
	return svc.STG.Upsert(ml.EntitySensorField, field, filters)
}

// GetByIDs returns a field details by gatewayID, nodeId, sensorID and fieldName of a message
func GetByIDs(gatewayID, nodeID, sensorID, fieldID string) (*fml.Field, error) {
	filters := []pml.Filter{
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
	filters := []pml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return svc.STG.Delete(ml.EntitySensorField, filters)
}

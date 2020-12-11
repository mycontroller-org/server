package field

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	stgml "github.com/mycontroller-org/backend/v2/pkg/model/storage"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
)

// List by filter and pagination
func List(f []pml.Filter, p *pml.Pagination) (*pml.Result, error) {
	out := make([]fml.Field, 0)
	return svc.STG.Find(ml.EntitySensorField, &out, f, p)
}

// Get returns a field
func Get(f []pml.Filter) (*fml.Field, error) {
	out := &fml.Field{}
	err := svc.STG.FindOne(ml.EntitySensorField, out, f)
	return out, err
}

// Save a field details
func Save(sensor *fml.Field) error {
	if sensor.ID == "" {
		sensor.ID = ut.RandUUID()
	}
	f := []pml.Filter{
		{Key: ml.KeyID, Value: sensor.ID},
	}
	return svc.STG.Upsert(ml.EntitySensorField, sensor, f)
}

// GetByIDs returns a field details by gatewayID, nodeId, sensorID and fieldName of a message
func GetByIDs(gatewayID, nodeID, sensorID, fieldID string) (*fml.Field, error) {
	f := []pml.Filter{
		{Key: ml.KeyGatewayID, Value: gatewayID},
		{Key: ml.KeyNodeID, Value: nodeID},
		{Key: ml.KeySensorID, Value: sensorID},
		{Key: ml.KeyFieldID, Value: fieldID},
	}
	out := &fml.Field{}
	err := svc.STG.FindOne(ml.EntitySensorField, out, f)
	return out, err
}

// Delete sensor fields
func Delete(IDs []string) (int64, error) {
	f := []pml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return svc.STG.Delete(ml.EntitySensorField, f)
}

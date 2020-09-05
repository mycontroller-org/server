package field

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	"github.com/mycontroller-org/backend/v2/pkg/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
)

// List by filter and pagination
func List(f []pml.Filter, p *pml.Pagination) ([]fml.Field, error) {
	out := make([]fml.Field, 0)
	svc.STG.Find(ml.EntitySensorField, f, p, &out)
	return out, nil
}

// Get returns a field
func Get(f []pml.Filter) (*fml.Field, error) {
	out := &fml.Field{}
	err := svc.STG.FindOne(ml.EntitySensorField, f, out)
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
	return svc.STG.Upsert(ml.EntitySensorField, f, sensor)
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
	err := svc.STG.FindOne(ml.EntitySensorField, f, out)
	return out, err
}

// Delete sensor fields
func Delete(IDs []string) (int64, error) {
	f := []pml.Filter{{Key: ml.KeyID, Operator: storage.OperatorIn, Value: IDs}}
	return svc.STG.Delete(ml.EntitySensorField, f)
}

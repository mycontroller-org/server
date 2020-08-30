package field

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
)

// List by filter and pagination
func List(f []pml.Filter, p pml.Pagination) ([]fml.Field, error) {
	out := make([]fml.Field, 0)
	svc.STG.Find(ml.EntityField, f, p, &out)
	return out, nil
}

// Get returns a field
func Get(f []pml.Filter) (*fml.Field, error) {
	out := &fml.Field{}
	err := svc.STG.FindOne(ml.EntityField, f, out)
	return out, err
}

// Save a field details
func Save(sensor *fml.Field) error {
	if sensor.ID == "" {
		sensor.ID = ut.RandID()
	}
	f := []pml.Filter{
		{Key: "id", Operator: "eq", Value: sensor.ID},
	}
	return svc.STG.Upsert(ml.EntityField, f, sensor)
}

// GetByIDs returns a field details by gatewayID, nodeId, sensorID and fieldName of a message
func GetByIDs(gatewayID, nodeID, sensorID, fieldName string) (*fml.Field, error) {
	id := fml.AssembleID(gatewayID, nodeID, sensorID, fieldName)
	f := []pml.Filter{
		{Key: "id", Operator: "eq", Value: id},
	}
	out := &fml.Field{}
	err := svc.STG.FindOne(ml.EntityField, f, out)
	return out, err
}

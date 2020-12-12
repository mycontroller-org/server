package sensor

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	sml "github.com/mycontroller-org/backend/v2/pkg/model/sensor"
	stgml "github.com/mycontroller-org/backend/v2/pkg/model/storage"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
)

// List by filter and pagination
func List(filters []pml.Filter, pagination *pml.Pagination) (*pml.Result, error) {
	result := make([]sml.Sensor, 0)
	return svc.STG.Find(ml.EntitySensor, &result, filters, pagination)
}

// Get returns a sensor
func Get(filters []pml.Filter) (*sml.Sensor, error) {
	result := &sml.Sensor{}
	err := svc.STG.FindOne(ml.EntitySensor, result, filters)
	return result, err
}

// Save a Sensor details
func Save(sensor *sml.Sensor) error {
	if sensor.ID == "" {
		sensor.ID = ut.RandUUID()
	}
	f := []pml.Filter{
		{Key: ml.KeyID, Value: sensor.ID},
	}
	return svc.STG.Upsert(ml.EntitySensor, sensor, f)
}

// GetByIDs returns a sensor details by gatewayID, nodeId and sensorID of a message
func GetByIDs(gatewayID, nodeID, sensorID string) (*sml.Sensor, error) {
	filters := []pml.Filter{
		{Key: ml.KeyGatewayID, Value: gatewayID},
		{Key: ml.KeyNodeID, Value: nodeID},
		{Key: ml.KeySensorID, Value: sensorID},
	}
	result := &sml.Sensor{}
	err := svc.STG.FindOne(ml.EntitySensor, result, filters)
	return result, err
}

// Delete sensor
func Delete(IDs []string) (int64, error) {
	filters := []pml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return svc.STG.Delete(ml.EntitySensor, filters)
}

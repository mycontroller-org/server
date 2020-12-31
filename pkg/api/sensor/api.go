package sensor

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	sml "github.com/mycontroller-org/backend/v2/pkg/model/sensor"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]sml.Sensor, 0)
	return stg.SVC.Find(ml.EntitySensor, &result, filters, pagination)
}

// Get returns a sensor
func Get(filters []stgml.Filter) (*sml.Sensor, error) {
	result := &sml.Sensor{}
	err := stg.SVC.FindOne(ml.EntitySensor, result, filters)
	return result, err
}

// Save a Sensor details
func Save(sensor *sml.Sensor) error {
	if sensor.ID == "" {
		sensor.ID = ut.RandUUID()
	}
	f := []stgml.Filter{
		{Key: ml.KeyID, Value: sensor.ID},
	}
	return stg.SVC.Upsert(ml.EntitySensor, sensor, f)
}

// GetByIDs returns a sensor details by gatewayID, nodeId and sensorID of a message
func GetByIDs(gatewayID, nodeID, sensorID string) (*sml.Sensor, error) {
	filters := []stgml.Filter{
		{Key: ml.KeyGatewayID, Value: gatewayID},
		{Key: ml.KeyNodeID, Value: nodeID},
		{Key: ml.KeySensorID, Value: sensorID},
	}
	result := &sml.Sensor{}
	err := stg.SVC.FindOne(ml.EntitySensor, result, filters)
	return result, err
}

// Delete sensor
func Delete(IDs []string) (int64, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(ml.EntitySensor, filters)
}

package sensor

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	sml "github.com/mycontroller-org/backend/v2/pkg/model/sensor"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	"github.com/mycontroller-org/backend/v2/pkg/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
)

// List by filter and pagination
func List(f []pml.Filter, p *pml.Pagination) ([]sml.Sensor, error) {
	out := make([]sml.Sensor, 0)
	svc.STG.Find(ml.EntitySensor, f, p, &out)
	return out, nil
}

// Get returns a sensor
func Get(f []pml.Filter) (*sml.Sensor, error) {
	out := &sml.Sensor{}
	err := svc.STG.FindOne(ml.EntitySensor, f, out)
	return out, err
}

// Save a Sensor details
func Save(sensor *sml.Sensor) error {
	if sensor.ID == "" {
		sensor.ID = ut.RandUUID()
	}
	f := []pml.Filter{
		{Key: ml.KeyID, Value: sensor.ID},
	}
	return svc.STG.Upsert(ml.EntitySensor, f, sensor)
}

// GetByIDs returns a sensor details by gatewayID, nodeId and sensorID of a message
func GetByIDs(gatewayID, nodeID, sensorID string) (*sml.Sensor, error) {
	f := []pml.Filter{
		{Key: ml.KeyGatewayID, Value: gatewayID},
		{Key: ml.KeyNodeID, Value: nodeID},
		{Key: ml.KeySensorID, Value: sensorID},
	}
	out := &sml.Sensor{}
	err := svc.STG.FindOne(ml.EntitySensor, f, out)
	return out, err
}

// Delete sensor
func Delete(IDs []string) (int64, error) {
	f := []pml.Filter{{Key: ml.KeyID, Operator: storage.OperatorIn, Value: IDs}}
	return svc.STG.Delete(ml.EntitySensor, f)
}

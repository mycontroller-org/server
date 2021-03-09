package action

import (
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// ToSensorFieldByID sends the payload to the given sensorfiled
func ToSensorFieldByID(id string, payload string) error {
	filters := []stgml.Filter{{Key: model.KeyID, Value: id}}
	field, err := fieldAPI.Get(filters)
	if err != nil {
		return err
	}
	return toSensorField(field.GatewayID, field.NodeID, field.SensorID, field.FieldID, payload)
}

func toSensorField(gatewayID, nodeID, sensorID, fieldID, payload string) error {
	if payload == model.ActionToggle {
		// get field current data
		field, err := fieldAPI.GetByIDs(gatewayID, nodeID, sensorID, fieldID)
		if err != nil {
			return err
		}

		if utils.ToBool(field.Current.Value) {
			payload = "false"
		} else {
			payload = "true"
		}
	}

	msg := msgml.NewMessage(false)
	msg.GatewayID = gatewayID
	msg.NodeID = nodeID
	msg.SensorID = sensorID
	pl := msgml.NewData()
	pl.Name = fieldID
	pl.Value = payload
	msg.Payloads = append(msg.Payloads, pl)
	msg.Type = msgml.TypeSet
	return Post(&msg)
}

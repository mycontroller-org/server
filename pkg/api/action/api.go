package action

import (
	"errors"
	"fmt"

	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// Execute the given request
func Execute(quickID, payload string) error {
	resource, kvMap, err := ml.ResourceKeyValueMap(quickID)
	if err != nil {
		return err
	}

	switch resource {
	case ml.QuickIDGateway:
	case ml.QuickIDNode:
	case ml.QuickIDNodeData:
	case ml.QuickIDSensor:
	case ml.QuickIDSensorField:
		msg := msgml.NewMessage(false)
		msg.GatewayID = kvMap[ml.KeyGatewayID]
		msg.NodeID = kvMap[ml.KeyNodeID]
		msg.SensorID = kvMap[ml.KeySensorID]
		pl := msgml.NewData()
		pl.Name = kvMap[ml.KeyFieldID]
		pl.Value = payload
		msg.Payloads = append(msg.Payloads, pl)
		msg.Type = msgml.TypeSet
		return Post(&msg)

	default:
		return fmt.Errorf("Unknown resource type: %s", resource)
	}
	return nil
}

// PostToSensorField sends the payload to the given sensorfiled
func PostToSensorField(id string, payload string) error {
	filters := []stgml.Filter{{Key: ml.KeyID, Value: id}}
	field, err := fieldAPI.Get(filters)
	if err != nil {
		return err
	}

	// send payload
	msg := msgml.NewMessage(false)
	msg.GatewayID = field.GatewayID
	msg.NodeID = field.NodeID
	msg.SensorID = field.SensorID
	pl := msgml.NewData()
	pl.Name = field.FieldID
	pl.Value = payload
	msg.Payloads = append(msg.Payloads, pl)
	msg.Type = msgml.TypeSet
	return Post(&msg)
}

// Post a message to gateway
func Post(msg *msgml.Message) error {
	if msg.GatewayID == "" {
		return errors.New("gateway id can not be empty")
	}
	topic := mcbus.GetTopicPostMessageToProvider(msg.GatewayID)
	return mcbus.Publish(topic, msg)
}

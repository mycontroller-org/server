package action

import (
	"errors"
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/mcbus"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	gwpd "github.com/mycontroller-org/backend/v2/plugin/gw_provider"
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

// Post a message to gateway
func Post(msg *msgml.Message) error {
	if msg.GatewayID == "" {
		return errors.New("gateway id can not be empty")
	}
	topic := gwpd.GetTopicListenFromCore(msg.GatewayID)
	return mcbus.Publish(topic, msg)
}

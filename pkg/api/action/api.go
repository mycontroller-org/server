package action

import (
	"errors"
	"fmt"

	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nodeml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	quickIdUL "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// ExecuteNodeAction for a node
func ExecuteNodeAction(action string, nodeIDs []string) error {
	// verify is a valid action?
	switch action {
	case nodeml.ActionDiscover,
		nodeml.ActionFirmwareUpdate,
		nodeml.ActionHeartbeatRequest,
		nodeml.ActionReboot,
		nodeml.ActionRefreshNodeInfo,
		nodeml.ActionReset:
		// nothing to do, just continue
	default:
		return fmt.Errorf("invalid node action:%s", action)
	}

	nodes, err := nodeAPI.GetByeIDs(nodeIDs)
	if err != nil {
		return err
	}
	for index := 0; index < len(nodes); index++ {
		node := nodes[index]
		msg := msgml.NewMessage(false)
		msg.GatewayID = node.GatewayID
		msg.NodeID = node.NodeID
		msg.Type = msgml.TypeAction
		pl := msgml.NewData()
		pl.Name = action
		pl.Value = ""
		msg.Payloads = append(msg.Payloads, pl)
		err = Post(&msg)
		if err != nil {
			return err
		}
	}
	return nil
}

// Execute the given request
func Execute(quickID, payload string) error {
	resource, kvMap, err := quickIdUL.ResourceKeyValueMap(quickID)
	if err != nil {
		return err
	}

	switch {
	case utils.ContainsString(quickIdUL.QuickIDGateway, resource):
	case utils.ContainsString(quickIdUL.QuickIDNode, resource):
	case utils.ContainsString(quickIdUL.QuickIDSensor, resource):
	case utils.ContainsString(quickIdUL.QuickIDSensorField, resource):
		msg := msgml.NewMessage(false)
		msg.GatewayID = kvMap[model.KeyGatewayID]
		msg.NodeID = kvMap[model.KeyNodeID]
		msg.SensorID = kvMap[model.KeySensorID]
		pl := msgml.NewData()
		pl.Name = kvMap[model.KeyFieldID]
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
	filters := []stgml.Filter{{Key: model.KeyID, Value: id}}
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

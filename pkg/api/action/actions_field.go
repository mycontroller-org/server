package action

import (
	types "github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	quickIdUtils "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
)

// sends the payload to the given field
func (a *ActionAPI) ToFieldByQuickID(quickID string, payload string) error {
	_, idsMap, err := quickIdUtils.EntityKeyValueMap(quickID)
	if err != nil {
		return err
	}

	// really needs to check these ids on internal database?
	field, err := a.api.Field().GetByIDs(idsMap[types.KeyGatewayID], idsMap[types.KeyNodeID], idsMap[types.KeySourceID], idsMap[types.KeyFieldID])
	if err != nil {
		return err
	}
	return a.toField(field.GatewayID, field.NodeID, field.SourceID, field.FieldID, payload)
}

// toField sends the payload to the given ids
func (a *ActionAPI) toField(gatewayID, nodeID, sourceID, fieldID, payload string) error {
	if payload == types.ActionToggle {
		// get field current data
		field, err := a.api.Field().GetByIDs(gatewayID, nodeID, sourceID, fieldID)
		if err != nil {
			return err
		}

		if converterUtils.ToBool(field.Current.Value) {
			payload = "false"
		} else {
			payload = "true"
		}
	}

	msg := msgTY.NewMessage(false)
	msg.GatewayID = gatewayID
	msg.NodeID = nodeID
	msg.SourceID = sourceID

	// get node details and update isPassiveNode
	node, err := a.api.Node().GetByGatewayAndNodeID(gatewayID, nodeID)
	if err == nil {
		msg.IsSleepNode = node.IsSleepNode()
	}

	pl := msgTY.NewPayload()
	pl.Key = fieldID
	pl.SetValue(payload)
	msg.Payloads = append(msg.Payloads, pl)
	msg.Type = msgTY.TypeSet
	return a.Post(&msg)
}

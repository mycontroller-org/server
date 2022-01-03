package action

import (
	fieldAPI "github.com/mycontroller-org/server/v2/pkg/api/field"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	quickIdUtils "github.com/mycontroller-org/server/v2/pkg/utils/quick_id"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// ToFieldByID sends the payload to the given field
func ToFieldByID(id string, payload string) error {
	filters := []storageTY.Filter{{Key: types.KeyID, Value: id}}
	field, err := fieldAPI.Get(filters)
	if err != nil {
		return err
	}
	return ToField(field.GatewayID, field.NodeID, field.SourceID, field.FieldID, payload)
}

// ToFieldByQuickID sends the payload to the given field
func ToFieldByQuickID(quickID string, payload string) error {
	_, idsMap, err := quickIdUtils.EntityKeyValueMap(quickID)
	if err != nil {
		return err
	}

	// really needs to check these ids on internal database?
	field, err := fieldAPI.GetByIDs(idsMap[types.KeyGatewayID], idsMap[types.KeyNodeID], idsMap[types.KeySourceID], idsMap[types.KeyFieldID])
	if err != nil {
		return err
	}
	return ToField(field.GatewayID, field.NodeID, field.SourceID, field.FieldID, payload)
}

// ToField sends the payload to the given ids
func ToField(gatewayID, nodeID, sourceID, fieldID, payload string) error {
	if payload == types.ActionToggle {
		// get field current data
		field, err := fieldAPI.GetByIDs(gatewayID, nodeID, sourceID, fieldID)
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
	pl := msgTY.NewPayload()
	pl.Key = fieldID
	pl.Value = payload
	msg.Payloads = append(msg.Payloads, pl)
	msg.Type = msgTY.TypeSet
	return Post(&msg)
}

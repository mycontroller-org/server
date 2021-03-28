package action

import (
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	quickIdUtils "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// ToFieldByID sends the payload to the given field
func ToFieldByID(id string, payload string) error {
	filters := []stgml.Filter{{Key: model.KeyID, Value: id}}
	field, err := fieldAPI.Get(filters)
	if err != nil {
		return err
	}
	return ToField(field.GatewayID, field.NodeID, field.SourceID, field.FieldID, payload)
}

// ToFieldByQuickID sends the payload to the given field
func ToFieldByQuickID(quickID string, payload string) error {
	_, idsMap, err := quickIdUtils.ResourceKeyValueMap(quickID)
	if err != nil {
		return err
	}

	// really needs to check these ids on internal database?
	field, err := fieldAPI.GetByIDs(idsMap[model.KeyGatewayID], idsMap[model.KeyNodeID], idsMap[model.KeySourceID], idsMap[model.KeyFieldID])
	if err != nil {
		return err
	}
	return ToField(field.GatewayID, field.NodeID, field.SourceID, field.FieldID, payload)
}

// ToField sends the payload to the given ids
func ToField(gatewayID, nodeID, sourceID, fieldID, payload string) error {
	if payload == model.ActionToggle {
		// get field current data
		field, err := fieldAPI.GetByIDs(gatewayID, nodeID, sourceID, fieldID)
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
	msg.SourceID = sourceID
	pl := msgml.NewData()
	pl.Name = fieldID
	pl.Value = payload
	msg.Payloads = append(msg.Payloads, pl)
	msg.Type = msgml.TypeSet
	return Post(&msg)
}

package field

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]fieldTY.Field, 0)
	return store.STORAGE.Find(types.EntityField, &result, filters, pagination)
}

// Get returns a field
func Get(filters []storageTY.Filter) (*fieldTY.Field, error) {
	result := &fieldTY.Field{}
	err := store.STORAGE.FindOne(types.EntityField, result, filters)
	return result, err
}

// GetByID returns a field
func GetByID(id string) (*fieldTY.Field, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	result := &fieldTY.Field{}
	err := store.STORAGE.FindOne(types.EntityField, result, filters)
	return result, err
}

// Save a field details
func Save(field *fieldTY.Field, retainValue bool) error {
	eventType := eventTY.TypeUpdated
	if field.ID == "" {
		field.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: field.ID},
	}

	if retainValue && eventType != eventTY.TypeCreated {
		fieldOrg, err := GetByID(field.ID)
		if err != nil {
			return err
		}
		field.Current = fieldOrg.Current
		field.Previous = fieldOrg.Previous
	}
	err := store.STORAGE.Upsert(types.EntityField, field, filters)
	if err != nil {
		return err
	}

	if retainValue { // assume this change from HTTP API
		busUtils.PostEvent(mcbus.TopicEventField, eventType, types.EntityField, field)
	}
	return nil
}

// GetByIDs returns a field details by gatewayID, nodeId, sourceID and fieldName of a message
func GetByIDs(gatewayID, nodeID, sourceID, fieldID string) (*fieldTY.Field, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyGatewayID, Value: gatewayID},
		{Key: types.KeyNodeID, Value: nodeID},
		{Key: types.KeySourceID, Value: sourceID},
		{Key: types.KeyFieldID, Value: fieldID},
	}
	result := &fieldTY.Field{}
	err := store.STORAGE.FindOne(types.EntityField, result, filters)
	return result, err
}

// Delete fields
func Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}

	fields := make([]fieldTY.Field, 0)
	pagination := &storageTY.Pagination{Limit: int64(len(IDs))}
	_, err := store.STORAGE.Find(types.EntityField, &fields, filters, pagination)
	if err != nil {
		return 0, err
	}
	deleted := int64(0)
	for _, field := range fields {
		deleteFilter := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorEqual, Value: field.ID}}
		_, err = store.STORAGE.Delete(types.EntityField, deleteFilter)
		if err != nil {
			return deleted, err
		}
		deleted++
		// post deletion event
		busUtils.PostEvent(mcbus.TopicEventField, eventTY.TypeDeleted, types.EntityField, field)
	}
	return deleted, nil
}

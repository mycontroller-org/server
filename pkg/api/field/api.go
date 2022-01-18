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
		busUtils.PostEvent(mcbus.TopicEventHandler, eventType, types.EntityHandler, field)
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
	return store.STORAGE.Delete(types.EntityField, filters)
}

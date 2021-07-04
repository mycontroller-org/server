package field

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
	fieldML "github.com/mycontroller-org/server/v2/pkg/model/field"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	stg "github.com/mycontroller-org/server/v2/pkg/service/database/storage"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// List by filter and pagination
func List(filters []stgType.Filter, pagination *stgType.Pagination) (*stgType.Result, error) {
	result := make([]fieldML.Field, 0)
	return stg.SVC.Find(model.EntityField, &result, filters, pagination)
}

// Get returns a field
func Get(filters []stgType.Filter) (*fieldML.Field, error) {
	result := &fieldML.Field{}
	err := stg.SVC.FindOne(model.EntityField, result, filters)
	return result, err
}

// GetByID returns a field
func GetByID(id string) (*fieldML.Field, error) {
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: id},
	}
	result := &fieldML.Field{}
	err := stg.SVC.FindOne(model.EntityField, result, filters)
	return result, err
}

// Save a field details
func Save(field *fieldML.Field, retainValue bool) error {
	eventType := eventML.TypeUpdated
	if field.ID == "" {
		field.ID = utils.RandUUID()
		eventType = eventML.TypeCreated
	}
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: field.ID},
	}

	if retainValue && eventType != eventML.TypeCreated {
		fieldOrg, err := GetByID(field.ID)
		if err != nil {
			return err
		}
		field.Current = fieldOrg.Current
		field.Previous = fieldOrg.Previous
	}
	err := stg.SVC.Upsert(model.EntityField, field, filters)
	if err != nil {
		return err
	}

	if retainValue { // assume this change from HTTP API
		busUtils.PostEvent(mcbus.TopicEventHandler, eventType, model.EntityHandler, field)
	}
	return nil
}

// GetByIDs returns a field details by gatewayID, nodeId, sourceID and fieldName of a message
func GetByIDs(gatewayID, nodeID, sourceID, fieldID string) (*fieldML.Field, error) {
	filters := []stgType.Filter{
		{Key: model.KeyGatewayID, Value: gatewayID},
		{Key: model.KeyNodeID, Value: nodeID},
		{Key: model.KeySourceID, Value: sourceID},
		{Key: model.KeyFieldID, Value: fieldID},
	}
	result := &fieldML.Field{}
	err := stg.SVC.FindOne(model.EntityField, result, filters)
	return result, err
}

// Delete fields
func Delete(IDs []string) (int64, error) {
	filters := []stgType.Filter{{Key: model.KeyID, Operator: stgType.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(model.EntityField, filters)
}

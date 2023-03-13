package field

import (
	"context"
	"fmt"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type FieldAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	bus     busTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, bus busTY.Plugin) *FieldAPI {
	return &FieldAPI{
		ctx:     ctx,
		logger:  logger.Named("field_api"),
		storage: storage,
		bus:     bus,
	}
}

// List by filter and pagination
func (f *FieldAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]fieldTY.Field, 0)
	return f.storage.Find(types.EntityField, &result, filters, pagination)
}

// Get returns a field
func (f *FieldAPI) Get(filters []storageTY.Filter) (*fieldTY.Field, error) {
	result := &fieldTY.Field{}
	err := f.storage.FindOne(types.EntityField, result, filters)
	return result, err
}

// GetByID returns a field
func (f *FieldAPI) GetByID(id string) (*fieldTY.Field, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	result := &fieldTY.Field{}
	err := f.storage.FindOne(types.EntityField, result, filters)
	return result, err
}

// Save a field details
func (f *FieldAPI) Save(field *fieldTY.Field, retainValue bool) error {
	eventType := eventTY.TypeUpdated
	if field.ID == "" {
		field.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: field.ID},
	}

	if retainValue && eventType != eventTY.TypeCreated {
		fieldOrg, err := f.GetByID(field.ID)
		if err != nil {
			return err
		}
		field.Current = fieldOrg.Current
		field.Previous = fieldOrg.Previous
	}
	err := f.storage.Upsert(types.EntityField, field, filters)
	if err != nil {
		return err
	}

	if retainValue { // assume this change from HTTP API
		busUtils.PostEvent(f.logger, f.bus, topic.TopicEventField, eventType, types.EntityField, field)
	}
	return nil
}

// GetByIDs returns a field details by gatewayID, nodeId, sourceID and fieldName of a message
func (f *FieldAPI) GetByIDs(gatewayID, nodeID, sourceID, fieldID string) (*fieldTY.Field, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyGatewayID, Value: gatewayID},
		{Key: types.KeyNodeID, Value: nodeID},
		{Key: types.KeySourceID, Value: sourceID},
		{Key: types.KeyFieldID, Value: fieldID},
	}
	result := &fieldTY.Field{}
	err := f.storage.FindOne(types.EntityField, result, filters)
	return result, err
}

// Delete fields
func (f *FieldAPI) Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}

	fields := make([]fieldTY.Field, 0)
	pagination := &storageTY.Pagination{Limit: int64(len(IDs))}
	_, err := f.storage.Find(types.EntityField, &fields, filters, pagination)
	if err != nil {
		return 0, err
	}
	deleted := int64(0)
	for _, field := range fields {
		deleteFilter := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorEqual, Value: field.ID}}
		_, err = f.storage.Delete(types.EntityField, deleteFilter)
		if err != nil {
			return deleted, err
		}
		deleted++
		// post deletion event
		busUtils.PostEvent(f.logger, f.bus, topic.TopicEventField, eventTY.TypeDeleted, types.EntityField, field)
	}
	return deleted, nil
}

func (f *FieldAPI) Import(data interface{}) error {
	input, ok := data.(fieldTY.Field)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		input.ID = utils.RandUUID()
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}

	return f.storage.Upsert(types.EntityField, &input, filters)
}

func (f *FieldAPI) GetEntityInterface() interface{} {
	return fieldTY.Field{}
}

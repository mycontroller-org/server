package forwardpayload

import (
	"context"
	"errors"
	"fmt"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	fwdPayloadTY "github.com/mycontroller-org/server/v2/pkg/types/forward_payload"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type ForwardPayloadAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	bus     busTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, bus busTY.Plugin) *ForwardPayloadAPI {
	return &ForwardPayloadAPI{
		ctx:     ctx,
		logger:  logger.Named("forward_payload_api"),
		storage: storage,
		bus:     bus,
	}
}

// List by filter and pagination
func (fpl *ForwardPayloadAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]fwdPayloadTY.Config, 0)
	return fpl.storage.Find(types.EntityForwardPayload, &result, filters, pagination)
}

// Get returns a item
func (fpl *ForwardPayloadAPI) Get(filters []storageTY.Filter) (*fwdPayloadTY.Config, error) {
	result := &fwdPayloadTY.Config{}
	err := fpl.storage.FindOne(types.EntityForwardPayload, result, filters)
	return result, err
}

// Save a item details
func (fpl *ForwardPayloadAPI) Save(fp *fwdPayloadTY.Config) error {
	eventType := eventTY.TypeUpdated
	if fp.ID == "" {
		fp.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: fp.ID},
	}
	err := fpl.storage.Upsert(types.EntityForwardPayload, fp, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(fpl.logger, fpl.bus, topic.TopicEventForwardPayload, eventType, types.EntityForwardPayload, fp)
	return nil
}

// Delete items
func (fpl *ForwardPayloadAPI) Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return fpl.storage.Delete(types.EntityForwardPayload, filters)
}

// Enable forward payload entries
func (fpl *ForwardPayloadAPI) Enable(ids []string) error {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	response, err := fpl.List(filters, pagination)
	if err != nil {
		return err
	}

	mappings := *response.Data.(*[]fwdPayloadTY.Config)
	for index := 0; index < len(mappings); index++ {
		mapping := mappings[index]
		if !mapping.Enabled {
			mapping.Enabled = true
			err = fpl.Save(&mapping)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Disable forward entries
func (fpl *ForwardPayloadAPI) Disable(ids []string) error {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	response, err := fpl.List(filters, pagination)
	if err != nil {
		return err
	}
	mappings := *response.Data.(*[]fwdPayloadTY.Config)
	for index := 0; index < len(mappings); index++ {
		mapping := mappings[index]
		if mapping.Enabled {
			mapping.Enabled = false
			err = fpl.Save(&mapping)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fpl *ForwardPayloadAPI) Import(data interface{}) error {
	input, ok := data.(fwdPayloadTY.Config)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		return errors.New("'id' can not be empty")
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	return fpl.storage.Upsert(types.EntityForwardPayload, &input, filters)
}

func (fpl *ForwardPayloadAPI) GetEntityInterface() interface{} {
	return fwdPayloadTY.Config{}
}

package forwardpayload

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	fwdPayloadTY "github.com/mycontroller-org/server/v2/pkg/types/forward_payload"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]fwdPayloadTY.Config, 0)
	return store.STORAGE.Find(types.EntityForwardPayload, &result, filters, pagination)
}

// Get returns a item
func Get(filters []storageTY.Filter) (*fwdPayloadTY.Config, error) {
	result := &fwdPayloadTY.Config{}
	err := store.STORAGE.FindOne(types.EntityForwardPayload, result, filters)
	return result, err
}

// Save a item details
func Save(fp *fwdPayloadTY.Config) error {
	eventType := eventTY.TypeUpdated
	if fp.ID == "" {
		fp.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: fp.ID},
	}
	err := store.STORAGE.Upsert(types.EntityForwardPayload, fp, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventForwardPayload, eventType, types.EntityForwardPayload, fp)
	return nil
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(types.EntityForwardPayload, filters)
}

// Enable forward payload entries
func Enable(ids []string) error {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	response, err := List(filters, pagination)
	if err != nil {
		return err
	}

	mappings := *response.Data.(*[]fwdPayloadTY.Config)
	for index := 0; index < len(mappings); index++ {
		mapping := mappings[index]
		if !mapping.Enabled {
			mapping.Enabled = true
			err = Save(&mapping)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Disable forward entries
func Disable(ids []string) error {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids}}
	pagination := &storageTY.Pagination{Limit: 100}
	response, err := List(filters, pagination)
	if err != nil {
		return err
	}
	mappings := *response.Data.(*[]fwdPayloadTY.Config)
	for index := 0; index < len(mappings); index++ {
		mapping := mappings[index]
		if mapping.Enabled {
			mapping.Enabled = false
			err = Save(&mapping)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

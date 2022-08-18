package virtual_device

import (
	"errors"
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]vdTY.VirtualDevice, 0)
	return store.STORAGE.Find(types.EntityVirtualDevice, &result, filters, pagination)
}

// returns a virtual device
func Get(filters []storageTY.Filter) (*vdTY.VirtualDevice, error) {
	result := &vdTY.VirtualDevice{}
	err := store.STORAGE.FindOne(types.EntityVirtualDevice, result, filters)
	return result, err
}

// GetByID returns a field
func GetByID(id string) (*vdTY.VirtualDevice, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	result := &vdTY.VirtualDevice{}
	err := store.STORAGE.FindOne(types.EntityVirtualDevice, result, filters)
	return result, err
}

// Save a virtual details
func Save(device *vdTY.VirtualDevice) error {
	if device.ID == "" {
		return errors.New("device id can not be empty")
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: device.ID},
	}

	// update quickId based resource details
	resources := []string{}
	for _, resource := range device.Traits {
		if resource.QuickID == "" {
			continue
		}
		resources = append(resources, fmt.Sprintf("%s:%s", resource.ResourceType, resource.QuickID))
	}
	device.Resources = resources

	err := store.STORAGE.Upsert(types.EntityVirtualDevice, device, filters)
	if err != nil {
		return err
	}

	return nil
}

// Delete virtual devices
func Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}

	devices := make([]vdTY.VirtualDevice, 0)
	pagination := &storageTY.Pagination{Limit: int64(len(IDs))}
	_, err := store.STORAGE.Find(types.EntityVirtualDevice, &devices, filters, pagination)
	if err != nil {
		return 0, err
	}
	deleted := int64(0)
	for _, device := range devices {
		deleteFilter := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorEqual, Value: device.ID}}
		_, err = store.STORAGE.Delete(types.EntityVirtualDevice, deleteFilter)
		if err != nil {
			return deleted, err
		}
		deleted++
		// post deletion event
		busUtils.PostEvent(mcbus.TopicEventVirtualDevice, eventTY.TypeDeleted, types.EntityVirtualDevice, device)
	}
	return deleted, nil
}

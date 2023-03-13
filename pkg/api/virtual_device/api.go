package virtual_device

import (
	"context"
	"errors"
	"fmt"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type VirtualDeviceAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	bus     busTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, bus busTY.Plugin) *VirtualDeviceAPI {
	return &VirtualDeviceAPI{
		ctx:     ctx,
		logger:  logger.Named("virtual_device_api"),
		storage: storage,
		bus:     bus,
	}
}

// List by filter and pagination
func (vd *VirtualDeviceAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]vdTY.VirtualDevice, 0)
	return vd.storage.Find(types.EntityVirtualDevice, &result, filters, pagination)
}

// returns a virtual device
func (vd *VirtualDeviceAPI) Get(filters []storageTY.Filter) (*vdTY.VirtualDevice, error) {
	result := &vdTY.VirtualDevice{}
	err := vd.storage.FindOne(types.EntityVirtualDevice, result, filters)
	return result, err
}

// GetByID returns a field
func (vd *VirtualDeviceAPI) GetByID(id string) (*vdTY.VirtualDevice, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	result := &vdTY.VirtualDevice{}
	err := vd.storage.FindOne(types.EntityVirtualDevice, result, filters)
	return result, err
}

// Save a virtual details
func (vd *VirtualDeviceAPI) Save(device *vdTY.VirtualDevice) error {
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

	err := vd.storage.Upsert(types.EntityVirtualDevice, device, filters)
	if err != nil {
		return err
	}

	return nil
}

// Delete virtual devices
func (vd *VirtualDeviceAPI) Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}

	devices := make([]vdTY.VirtualDevice, 0)
	pagination := &storageTY.Pagination{Limit: int64(len(IDs))}
	_, err := vd.storage.Find(types.EntityVirtualDevice, &devices, filters, pagination)
	if err != nil {
		return 0, err
	}
	deleted := int64(0)
	for _, device := range devices {
		deleteFilter := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorEqual, Value: device.ID}}
		_, err = vd.storage.Delete(types.EntityVirtualDevice, deleteFilter)
		if err != nil {
			return deleted, err
		}
		deleted++
		// post deletion event
		busUtils.PostEvent(vd.logger, vd.bus, topic.TopicEventVirtualDevice, eventTY.TypeDeleted, types.EntityVirtualDevice, device)
	}
	return deleted, nil
}

func (vd *VirtualDeviceAPI) Import(data interface{}) error {
	input, ok := data.(vdTY.VirtualDevice)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		input.ID = utils.RandUUID()
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	return vd.storage.Upsert(types.EntityVirtualDevice, &input, filters)
}

func (vd *VirtualDeviceAPI) GetEntityInterface() interface{} {
	return vdTY.VirtualDevice{}
}

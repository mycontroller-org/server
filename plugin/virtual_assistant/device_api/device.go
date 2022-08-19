package api

import (
	"errors"
	"fmt"
	"strings"
	"time"

	quickIdAPI "github.com/mycontroller-org/server/v2/pkg/api/quickid"
	vdAPI "github.com/mycontroller-org/server/v2/pkg/api/virtual_device"
	"github.com/mycontroller-org/server/v2/pkg/types"
	filedTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	filterUtil "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

const (
	DefaultLimit  = 500
	DefaultOffset = 0
)

func ListDevices(filters []storageTY.Filter, limit, offset int64) ([]vdTY.VirtualDevice, error) {
	if filters == nil {
		filters = make([]storageTY.Filter, 0)
	} else {
		// remove enabled filter
		_filters := make([]storageTY.Filter, 0)
		for _, filter := range filters { // removes "enabled" filter
			if !strings.EqualFold(filter.Key, types.KeyEnabled) {
				_filters = append(_filters, filter)
			}
		}
		filters = _filters
	}

	// add enabled filter
	filters = append(filters, storageTY.Filter{Key: types.KeyEnabled, Operator: storageTY.OperatorEqual, Value: true})

	pagination := &storageTY.Pagination{
		Offset: offset,
		Limit:  limit,
		SortBy: []storageTY.Sort{
			{Field: types.KeyLocation, OrderBy: storageTY.SortByASC},
			{Field: types.KeyName, OrderBy: storageTY.SortByASC},
		},
	}

	result, err := vdAPI.List(filters, pagination)
	if err != nil {
		zap.L().Error("error on getting devices", zap.Error(err))
		return nil, err
	}

	if result.Count > 0 {
		vDevicesPointer, ok := result.Data.(*[]vdTY.VirtualDevice)
		if ok {
			vDevices := *vDevicesPointer
			return vDevices, nil
		} else {
			zap.L().Warn("error on type casting", zap.String("received", fmt.Sprintf("%T", result.Data)))
		}
	}

	devices := make([]vdTY.VirtualDevice, 0)
	return devices, nil
}

func UpdateDeviceState(vDevices []vdTY.VirtualDevice) error {
	for _, vDevice := range vDevices {
		for trait := range vDevice.Traits {
			vResource := vDevice.Traits[trait]
			value, valueTimestamp, err := GetResourceState(&vDevice, trait, &vResource)
			if err != nil {
				return err
			}
			vResource.Value = value
			vResource.ValueTimestamp = valueTimestamp
			vDevice.Traits[trait] = vResource // in map we are not getting the reference, hence replace with the original resource
		}
	}
	return nil
}

func GetResourceState(device *vdTY.VirtualDevice, trait string, vResource *vdTY.Resource) (interface{}, time.Time, error) {
	valueTimestamp := time.Time{}

	if vResource.Type != vdTY.ResourceByQuickID {
		return nil, valueTimestamp, errors.New("label based resources are not allowed")
	}
	quickID := fmt.Sprintf("%s:%s", vResource.ResourceType, vResource.QuickID)
	responseMap, err := quickIdAPI.GetResources([]string{quickID})
	if err != nil {
		return nil, valueTimestamp, err
	}
	resource, ok := responseMap[quickID]
	if !ok {
		return nil, valueTimestamp, fmt.Errorf("specified quickId is not found. quickId:%s, deviceId:%s, deviceName:%s", quickID, device.ID, device.Name)
	}

	keyPath := ""
	timestampPath := ""
	switch resource.(type) {
	case *filedTY.Field:
		keyPath = "current.value"
		timestampPath = "current.timestamp"

	default:
	}

	if keyPath == "" {
		return nil, valueTimestamp, fmt.Errorf("support not implemented for this resource type[%T]", resource)
	}

	// fetch value
	_, value, err := filterUtil.GetValueByKeyPath(resource, keyPath)
	if err != nil {
		zap.L().Error("error on getting a value on a resource", zap.String("trait", trait), zap.String("keyPath", keyPath), zap.Error(err))
		return nil, valueTimestamp, err
	}

	// fetch timestamp
	if timestampPath != "" {
		_, rawTimestamp, err := filterUtil.GetValueByKeyPath(resource, timestampPath)
		if err != nil {
			zap.L().Error("error on getting a timestamp on a resource", zap.String("trait", trait), zap.String("timestampPath", timestampPath), zap.Error(err))
		} else {
			if timestamp, ok := rawTimestamp.(time.Time); ok {
				valueTimestamp = timestamp
			}
		}
	}

	return value, valueTimestamp, nil
}

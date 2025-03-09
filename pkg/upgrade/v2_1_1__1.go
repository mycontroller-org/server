package upgrade

import (
	"context"
	"errors"
	"fmt"
	"time"

	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	backupTY "github.com/mycontroller-org/server/v2/plugin/database/storage/backup"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// from version 2.1.1, there is a schema changes on the virtual_device
// hence, here handling schema to restore from previous version of server backup
// also handles data migration from previous version of server

// virtualDevice struct
type virtualDeviceOld_2_1_0 struct {
	ID          string                       `json:"id" yaml:"id"`
	Name        string                       `json:"name" yaml:"name"`
	Description string                       `json:"description" yaml:"description"`
	Enabled     bool                         `json:"enabled" yaml:"enabled"`
	DeviceType  string                       `json:"deviceType" yaml:"deviceType"`
	Traits      map[string]resourceOld_2_1_0 `json:"traits" yaml:"traits"`
	Location    string                       `json:"location" yaml:"location"`
	Labels      cmap.CustomStringMap         `json:"labels" yaml:"labels"`
	ModifiedOn  time.Time                    `json:"modifiedOn" yaml:"modifiedOn"`
	Resources   []string                     `json:"resources" yaml:"resources"`
}

type resourceOld_2_1_0 struct {
	Type           string               `json:"type" yaml:"type"`
	ResourceType   string               `json:"resourceType" yaml:"resourceType"`
	QuickID        string               `json:"quickId" yaml:"quickId"`
	Labels         cmap.CustomStringMap `json:"labels" yaml:"labels"`
	Value          interface{}          `json:"-" yaml:"-"`
	ValueTimestamp time.Time            `json:"-" yaml:"-"`
}

// there is a change in virtual device traits structure
// get the existing virtual devices and update to new structure
func upgrade_2_1_1__1(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, api *entitiesAPI.API) error {
	filters := []storageTY.Filter{}
	pagination := &storageTY.Pagination{}

	recordLimit := int64(20)
	offset := int64(0)
	for {
		pagination.Offset = offset
		pagination.Limit = recordLimit
		vDevices := make([]virtualDeviceOld_2_1_0, 0)
		result, err := storage.Find(types.EntityVirtualDevice, &vDevices, filters, pagination)
		if err != nil {
			logger.Error("error on getting virtual_devices", zap.Error(err))
			return err
		}
		data, ok := result.Data.(*[]virtualDeviceOld_2_1_0)
		if !ok {
			logger.Error("received invalid type", zap.String("actualType", fmt.Sprintf("%T", result.Data)))
			return errors.New("received invalid type")
		}

		for _, vDevice := range *data {
			loggerWithEntity := logger.With(zap.String("entityName", fmt.Sprintf("virtual_device:%s", vDevice.ID)))

			// create updated
			updated := vdTY.VirtualDevice{
				ID:          vDevice.ID,
				Name:        vDevice.Name,
				Description: vDevice.Description,
				Enabled:     vDevice.Enabled,
				DeviceType:  vDevice.DeviceType,
				Location:    vDevice.Location,
				Labels:      vDevice.Labels,
				ModifiedOn:  vDevice.ModifiedOn,
				Resources:   vDevice.Resources,
			}

			// update traits
			traits := []vdTY.Resource{}
			for key, value := range vDevice.Traits {
				if value.Type != "resource_by_quick_id" {
					loggerWithEntity.Warn("only quick id based traits supported, removing the selected traits", zap.String("deviceId", vDevice.ID), zap.String("deviceName", vDevice.Name), zap.Any("trait", key), zap.Any("traitData", value))
					continue
				}
				trait := vdTY.Resource{
					Name:         key,
					TraitType:    key,
					ResourceType: value.ResourceType,
					QuickID:      value.QuickID,
					Labels:       value.Labels,
				}
				traits = append(traits, trait)
			}
			updated.Traits = traits

			// save the changes
			err = api.VirtualDevice().Save(&updated)
			if err != nil {
				loggerWithEntity.Error("error on saving a virtual device", zap.Any("device", updated), zap.Error(err))
				return err
			}
		}

		offset += recordLimit
		if result.Count < offset {
			break
		}
	}

	return nil
}

func updateRestoreApiMap_2_1_1(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, apiMap map[string]backupTY.Backup) (map[string]backupTY.Backup, error) {
	apiMap[types.EntityVirtualDevice] = &virtualDeviceApiOld_2_1_0{
		ctx:     ctx,
		logger:  logger,
		storage: storage,
	}
	return apiMap, nil
}

type virtualDeviceApiOld_2_1_0 struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
}

func (vd *virtualDeviceApiOld_2_1_0) Import(data interface{}) error {
	input, ok := data.(virtualDeviceOld_2_1_0)
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

func (vd *virtualDeviceApiOld_2_1_0) GetEntityInterface() interface{} {
	return virtualDeviceOld_2_1_0{}
}

func (vd *virtualDeviceApiOld_2_1_0) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	return nil, errors.New("this method not implemented")
}

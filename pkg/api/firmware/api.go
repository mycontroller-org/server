package firmware

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]fml.Firmware, 0)
	return stg.SVC.Find(ml.EntityFirmware, &result, filters, pagination)
}

// Get returns a item
func Get(filters []stgml.Filter) (fml.Firmware, error) {
	result := fml.Firmware{}
	err := stg.SVC.FindOne(ml.EntityFirmware, &result, filters)
	return result, err
}

// GetByID returns a firmware details by ID
func GetByID(id string) (fml.Firmware, error) {
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: id},
	}
	result := fml.Firmware{}
	err := stg.SVC.FindOne(ml.EntityFirmware, &result, filters)
	return result, err
}

// Save config into disk
func Save(firmware *fml.Firmware) error {
	if firmware.ID == "" {
		firmware.ID = ut.RandID()
	}
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: firmware.ID},
	}
	return stg.SVC.Upsert(ml.EntityFirmware, firmware, filters)
}

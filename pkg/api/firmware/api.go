package firmware

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
)

// List by filter and pagination
func List(filters []pml.Filter, pagination *pml.Pagination) (*pml.Result, error) {
	result := make([]fml.Firmware, 0)
	return svc.STG.Find(ml.EntityFirmware, &result, filters, pagination)
}

// Get returns a item
func Get(filters []pml.Filter) (fml.Firmware, error) {
	result := fml.Firmware{}
	err := svc.STG.FindOne(ml.EntityFirmware, &result, filters)
	return result, err
}

// GetByID returns a firmware details by ID
func GetByID(id string) (fml.Firmware, error) {
	filters := []pml.Filter{
		{Key: ml.KeyID, Value: id},
	}
	result := fml.Firmware{}
	err := svc.STG.FindOne(ml.EntityFirmware, &result, filters)
	return result, err
}

// Save config into disk
func Save(firmware *fml.Firmware) error {
	if firmware.ID == "" {
		firmware.ID = ut.RandID()
	}
	filters := []pml.Filter{
		{Key: ml.KeyID, Value: firmware.ID},
	}
	return svc.STG.Upsert(ml.EntityFirmware, firmware, filters)
}

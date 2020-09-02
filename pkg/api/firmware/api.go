package firmware

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
)

// List by filter and pagination
func List(f []pml.Filter, p *pml.Pagination) ([]fml.Firmware, error) {
	out := make([]fml.Firmware, 0)
	svc.STG.Find(ml.EntityFirmware, f, p, &out)
	return out, nil
}

// Get returns a item
func Get(f []pml.Filter) (fml.Firmware, error) {
	out := fml.Firmware{}
	err := svc.STG.FindOne(ml.EntityFirmware, f, &out)
	return out, err
}

// GetByID returns a firmware details by ID
func GetByID(id string) (fml.Firmware, error) {
	f := []pml.Filter{
		{Key: "id", Operator: "eq", Value: id},
	}
	out := fml.Firmware{}
	err := svc.STG.FindOne(ml.EntityFirmware, f, &out)
	return out, err
}

// Save config into disk
func Save(fw *fml.Firmware) error {
	if fw.ID == "" {
		fw.ID = ut.RandID()
	}
	f := []pml.Filter{
		{Key: "id", Operator: "eq", Value: fw.ID},
	}
	return svc.STG.Upsert(ml.EntityFirmware, f, fw)
}

package firmware

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/firmware"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
)

// List by filter and pagination
func List(f []pml.Filter, p *pml.Pagination) (*pml.Result, error) {
	out := make([]fml.Firmware, 0)
	return svc.STG.Find(ml.EntityFirmware, &out, f, p)
}

// Get returns a item
func Get(f []pml.Filter) (fml.Firmware, error) {
	out := fml.Firmware{}
	err := svc.STG.FindOne(ml.EntityFirmware, &out, f)
	return out, err
}

// GetByID returns a firmware details by ID
func GetByID(id string) (fml.Firmware, error) {
	f := []pml.Filter{
		{Key: ml.KeyID, Value: id},
	}
	out := fml.Firmware{}
	err := svc.STG.FindOne(ml.EntityFirmware, &out, f)
	return out, err
}

// Save config into disk
func Save(fw *fml.Firmware) error {
	if fw.ID == "" {
		fw.ID = ut.RandID()
	}
	f := []pml.Filter{
		{Key: ml.KeyID, Value: fw.ID},
	}
	return svc.STG.Upsert(ml.EntityFirmware, fw, f)
}

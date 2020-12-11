package gateway

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	stgml "github.com/mycontroller-org/backend/v2/pkg/model/storage"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
)

// List by filter and pagination
func List(f []pml.Filter, p *pml.Pagination) (*pml.Result, error) {
	out := make([]gwml.Config, 0)
	return svc.STG.Find(ml.EntityGateway, &out, f, p)
}

// Get returns a gateway
func Get(f []pml.Filter) (gwml.Config, error) {
	out := gwml.Config{}
	err := svc.STG.FindOne(ml.EntityGateway, &out, f)
	return out, err
}

// Save gateway config into disk
func Save(gwCfg *gwml.Config) error {
	if gwCfg.ID == "" {
		gwCfg.ID = ut.RandID()
	}
	return svc.STG.Upsert(ml.EntityGateway, gwCfg, nil)
}

// SetState Updates state data
func SetState(gwCfg *gwml.Config, s ml.State) error {
	gwCfg.State = s
	return Save(gwCfg)
}

// Delete gateway
func Delete(IDs []string) error {
	f := []pml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	svc.STG.Delete(ml.EntityGateway, f)
	return nil
}

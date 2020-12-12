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
func List(filters []pml.Filter, pagination *pml.Pagination) (*pml.Result, error) {
	result := make([]gwml.Config, 0)
	return svc.STG.Find(ml.EntityGateway, &result, filters, pagination)
}

// Get returns a gateway
func Get(filters []pml.Filter) (gwml.Config, error) {
	result := gwml.Config{}
	err := svc.STG.FindOne(ml.EntityGateway, &result, filters)
	return result, err
}

// Save gateway config into disk
func Save(gwCfg *gwml.Config) error {
	if gwCfg.ID == "" {
		gwCfg.ID = ut.RandID()
	}
	return svc.STG.Upsert(ml.EntityGateway, gwCfg, nil)
}

// SetState Updates state data
func SetState(gwCfg *gwml.Config, state ml.State) error {
	gwCfg.State = state
	return Save(gwCfg)
}

// Delete gateway
func Delete(IDs []string) error {
	filters := []pml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	svc.STG.Delete(ml.EntityGateway, filters)
	return nil
}

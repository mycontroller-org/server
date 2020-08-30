package gateway

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
)

// List by filter and pagination
func List(f []pml.Filter, p pml.Pagination) ([]gwml.Config, error) {
	out := make([]gwml.Config, 0)
	svc.STG.Find(ml.EntityGateway, f, p, &out)
	return out, nil
}

// Get returns a gateway
func Get(f []pml.Filter) (gwml.Config, error) {
	out := gwml.Config{}
	err := svc.STG.FindOne(ml.EntityGateway, f, &out)
	return out, err
}

// Save gateway config into disk
func Save(gwCfg *gwml.Config) error {
	if gwCfg.ID == "" {
		gwCfg.ID = ut.RandID()
	}
	return svc.STG.Upsert(ml.EntityGateway, nil, gwCfg)
}

// SetState Updates state data
func SetState(gwCfg *gwml.Config, s ml.State) error {
	gwCfg.State = s
	return Save(gwCfg)
}

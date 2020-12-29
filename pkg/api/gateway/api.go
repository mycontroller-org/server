package gateway

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]gwml.Config, 0)
	return svc.STG.Find(ml.EntityGateway, &result, filters, pagination)
}

// Get returns a gateway
func Get(filters []stgml.Filter) (gwml.Config, error) {
	result := gwml.Config{}
	err := svc.STG.FindOne(ml.EntityGateway, &result, filters)
	return result, err
}

// SaveAndReload gateway
func SaveAndReload(gwCfg *gwml.Config) error {
	err := Save(gwCfg)
	if err != nil {
		return err
	}
	return Reload(gwCfg.ID)
}

// Save gateway config
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
func Delete(IDs []string) (int64, error) {
	for _, id := range IDs {
		err := Disable(id)
		if err != nil {
			return 0, err
		}
	}
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return svc.STG.Delete(ml.EntityGateway, filters)
}

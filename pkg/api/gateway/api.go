package gateway

import (
	"github.com/mycontroller-org/backend/v2/pkg/model"
	eventML "github.com/mycontroller-org/backend/v2/pkg/model/bus/event"
	gwML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	cloneUtil "github.com/mycontroller-org/backend/v2/pkg/utils/clone"
	stgML "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgML.Filter, pagination *stgML.Pagination) (*stgML.Result, error) {
	result := make([]gwML.Config, 0)
	return stg.SVC.Find(model.EntityGateway, &result, filters, pagination)
}

// Get returns a gateway
func Get(filters []stgML.Filter) (*gwML.Config, error) {
	result := &gwML.Config{}
	err := stg.SVC.FindOne(model.EntityGateway, result, filters)
	return result, err
}

// GetByID returns a gateway details
func GetByID(id string) (*gwML.Config, error) {
	filters := []stgML.Filter{
		{Key: model.KeyID, Value: id},
	}
	result := &gwML.Config{}
	err := stg.SVC.FindOne(model.EntityGateway, result, filters)
	return result, err
}

// SaveAndReload gateway
func SaveAndReload(gwCfg *gwML.Config) error {
	gwCfg.State = &model.State{} //reset state
	err := Save(gwCfg)
	if err != nil {
		return err
	}
	return Reload([]string{gwCfg.ID})
}

// Save gateway config
func Save(gwCfg *gwML.Config) error {
	eventType := eventML.TypeUpdated
	if gwCfg.ID == "" {
		gwCfg.ID = utils.RandID()
		eventType = eventML.TypeCreated
	}

	// encrypt passwords, tokens
	err := cloneUtil.UpdateSecrets(gwCfg, true)
	if err != nil {
		return err
	}

	err = stg.SVC.Upsert(model.EntityGateway, gwCfg, nil)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventGateway, eventType, model.EntityGateway, gwCfg)
	return nil
}

// SetState Updates state data
func SetState(id string, state *model.State) error {
	gwCfg, err := GetByID(id)
	if err != nil {
		return err
	}
	gwCfg.State = state
	return Save(gwCfg)
}

// Delete gateway
func Delete(ids []string) (int64, error) {
	err := Disable(ids)
	if err != nil {
		return 0, err
	}
	filters := []stgML.Filter{{Key: model.KeyID, Operator: stgML.OperatorIn, Value: ids}}
	return stg.SVC.Delete(model.EntityGateway, filters)
}

package gateway

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
	gwType "github.com/mycontroller-org/server/v2/plugin/gateway/type"
)

// List by filter and pagination
func List(filters []stgType.Filter, pagination *stgType.Pagination) (*stgType.Result, error) {
	result := make([]gwType.Config, 0)
	return store.STORAGE.Find(model.EntityGateway, &result, filters, pagination)
}

// Get returns a gateway
func Get(filters []stgType.Filter) (*gwType.Config, error) {
	result := &gwType.Config{}
	err := store.STORAGE.FindOne(model.EntityGateway, result, filters)
	return result, err
}

// GetByIDs returns a gateway details by id
func GetByIDs(ids []string) ([]gwType.Config, error) {
	filters := []stgType.Filter{
		{Key: model.KeyID, Operator: stgType.OperatorIn, Value: ids},
	}
	pagination := &stgType.Pagination{Limit: int64(len(ids))}
	gateways := make([]gwType.Config, 0)
	_, err := store.STORAGE.Find(model.EntityNode, &gateways, filters, pagination)
	return gateways, err
}

// GetByID returns a gateway details
func GetByID(id string) (*gwType.Config, error) {
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: id},
	}
	result := &gwType.Config{}
	err := store.STORAGE.FindOne(model.EntityGateway, result, filters)
	return result, err
}

// SaveAndReload gateway
func SaveAndReload(gwCfg *gwType.Config) error {
	gwCfg.State = &model.State{} //reset state
	err := Save(gwCfg)
	if err != nil {
		return err
	}
	return Reload([]string{gwCfg.ID})
}

// Save gateway config
func Save(gwCfg *gwType.Config) error {
	eventType := eventML.TypeUpdated
	if gwCfg.ID == "" {
		gwCfg.ID = utils.RandID()
		eventType = eventML.TypeCreated
	}

	// encrypt passwords, tokens
	err := cloneUtil.UpdateSecrets(gwCfg, store.CFG.Secret, true)
	if err != nil {
		return err
	}

	err = store.STORAGE.Upsert(model.EntityGateway, gwCfg, nil)
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
	filters := []stgType.Filter{{Key: model.KeyID, Operator: stgType.OperatorIn, Value: ids}}
	return store.STORAGE.Delete(model.EntityGateway, filters)
}

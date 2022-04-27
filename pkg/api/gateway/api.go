package gateway

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]gwTY.Config, 0)
	return store.STORAGE.Find(types.EntityGateway, &result, filters, pagination)
}

// Get returns a gateway
func Get(filters []storageTY.Filter) (*gwTY.Config, error) {
	result := &gwTY.Config{}
	err := store.STORAGE.FindOne(types.EntityGateway, result, filters)
	return result, err
}

// GetByIDs returns a gateway details by id
func GetByIDs(ids []string) ([]gwTY.Config, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids},
	}
	pagination := &storageTY.Pagination{Limit: int64(len(ids))}
	gateways := make([]gwTY.Config, 0)
	_, err := store.STORAGE.Find(types.EntityGateway, &gateways, filters, pagination)
	return gateways, err
}

// GetByID returns a gateway details
func GetByID(id string) (*gwTY.Config, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	result := &gwTY.Config{}
	err := store.STORAGE.FindOne(types.EntityGateway, result, filters)
	return result, err
}

// SaveAndReload gateway
func SaveAndReload(gwCfg *gwTY.Config) error {
	gwCfg.State = &types.State{} //reset state
	err := Save(gwCfg)
	if err != nil {
		return err
	}
	return Reload([]string{gwCfg.ID})
}

// Save gateway config
func Save(gwCfg *gwTY.Config) error {
	eventType := eventTY.TypeUpdated
	if gwCfg.ID == "" {
		gwCfg.ID = utils.RandID()
		eventType = eventTY.TypeCreated
	}

	// encrypt passwords, tokens
	err := cloneUtil.UpdateSecrets(gwCfg, store.CFG.Secret, "", true, cloneUtil.DefaultSpecialKeys)
	if err != nil {
		return err
	}

	err = store.STORAGE.Upsert(types.EntityGateway, gwCfg, nil)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventGateway, eventType, types.EntityGateway, gwCfg)
	return nil
}

// SetState Updates state data
func SetState(id string, state *types.State) error {
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

	// delete one by one and send deletion event
	gateways, err := GetByIDs(ids)
	if err != nil {
		return 0, err
	}
	deleted := int64(0)
	for _, gateway := range gateways {
		filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorEqual, Value: gateway.ID}}
		_, err = store.STORAGE.Delete(types.EntityGateway, filters)
		if err != nil {
			return deleted, err
		}
		deleted++
		// deletion event
		busUtils.PostEvent(mcbus.TopicEventGateway, eventTY.TypeDeleted, types.EntityGateway, gateway)
	}

	return deleted, nil
}

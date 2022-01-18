package datarepository

import (
	"errors"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/service/configuration"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	repositoryTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]repositoryTY.Config, 0)
	return store.STORAGE.Find(types.EntityDataRepository, &result, filters, pagination)
}

// Get returns a item
func Get(filters []storageTY.Filter) (*repositoryTY.Config, error) {
	result := &repositoryTY.Config{}
	err := store.STORAGE.FindOne(types.EntityDataRepository, result, filters)
	if err == nil {
		updateResult, err := updateResult(result)
		if err != nil {
			return nil, err
		}
		result = updateResult
	}
	return result, err
}

// Save is used to update items from UI
func Save(data *repositoryTY.Config) error {
	if data.ID == "" {
		return errors.New("'id' can not be empty")
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: data.ID},
	}

	if !configuration.PauseModifiedOnUpdate.IsSet() {
		data.ModifiedOn = time.Now()
	}

	// encrypt passwords, tokens
	err := cloneUtil.UpdateSecrets(data, store.CFG.Secret, "", true, cloneUtil.DefaultSpecialKeys)
	if err != nil {
		return err
	}

	// in mongodb can not save map[interface{}]interface{} type
	// convert it to map[string]interface{} type
	if store.CFG.Database.Storage.GetString(types.KeyType) == storageTY.TypeMongoDB {
		updatedResult, err := updateResult(data)
		if err != nil {
			return err
		}
		data = updatedResult
	}

	err = store.STORAGE.Upsert(types.EntityDataRepository, data, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventDataRepository, eventTY.TypeUpdated, types.EntityDataRepository, data)
	return nil
}

// GetByID returns a item by id
func GetByID(id string) (*repositoryTY.Config, error) {
	f := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	out := &repositoryTY.Config{}
	err := store.STORAGE.FindOne(types.EntityDataRepository, out, f)
	if err == nil {
		updatedResult, err := updateResult(out)
		if err != nil {
			return nil, err
		}
		out = updatedResult
	}
	return out, err
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(types.EntityDataRepository, filters)
}

// map[interface{}]interface{} type not working as expected in javascript in task module
// convert it to map[string]interface{}, by calling json Marshal and Unmarshal
func updateResult(data *repositoryTY.Config) (*repositoryTY.Config, error) {
	updateResult := &repositoryTY.Config{}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(dataBytes, updateResult)
	if err != nil {
		return nil, err
	}
	return updateResult, nil
}

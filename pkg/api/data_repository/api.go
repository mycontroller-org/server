package datarepository

import (
	"errors"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/model"
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
	repositoryML "github.com/mycontroller-org/server/v2/pkg/model/data_repository"
	"github.com/mycontroller-org/server/v2/pkg/service/configuration"
	stg "github.com/mycontroller-org/server/v2/pkg/service/database/storage"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// List by filter and pagination
func List(filters []stgType.Filter, pagination *stgType.Pagination) (*stgType.Result, error) {
	result := make([]repositoryML.Config, 0)
	return stg.SVC.Find(model.EntityDataRepository, &result, filters, pagination)
}

// Get returns a item
func Get(filters []stgType.Filter) (*repositoryML.Config, error) {
	result := &repositoryML.Config{}
	err := stg.SVC.FindOne(model.EntityDataRepository, result, filters)
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
func Save(data *repositoryML.Config) error {
	if data.ID == "" {
		return errors.New("'id' can not be empty")
	}
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: data.ID},
	}

	if !configuration.PauseModifiedOnUpdate.IsSet() {
		data.ModifiedOn = time.Now()
	}

	// encrypt passwords, tokens
	err := cloneUtil.UpdateSecrets(data, true)
	if err != nil {
		return err
	}

	// in mongodb can not save map[interface{}]interface{} type
	// convert it to map[string]interface{} type
	if configuration.CFG.Database.Storage.GetString(model.KeyType) == stgType.TypeMongoDB {
		updatedResult, err := updateResult(data)
		if err != nil {
			return err
		}
		data = updatedResult
	}

	err = stg.SVC.Upsert(model.EntityDataRepository, data, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventDataRepository, eventML.TypeUpdated, model.EntityDataRepository, data)
	return nil
}

// GetByID returns a item by id
func GetByID(id string) (*repositoryML.Config, error) {
	f := []stgType.Filter{
		{Key: model.KeyID, Value: id},
	}
	out := &repositoryML.Config{}
	err := stg.SVC.FindOne(model.EntityDataRepository, out, f)
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
	filters := []stgType.Filter{{Key: model.KeyID, Operator: stgType.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(model.EntityDataRepository, filters)
}

// map[interface{}]interface{} type not working as expected in javascript in task module
// convert it to map[string]interface{}, by calling json Marshal and Unmarshal
func updateResult(data *repositoryML.Config) (*repositoryML.Config, error) {
	updateResult := &repositoryML.Config{}
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

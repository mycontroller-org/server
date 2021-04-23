package datarepository

import (
	"errors"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	eventML "github.com/mycontroller-org/backend/v2/pkg/model/bus/event"
	repositoryML "github.com/mycontroller-org/backend/v2/pkg/model/data_repository"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	stgML "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgML.Filter, pagination *stgML.Pagination) (*stgML.Result, error) {
	result := make([]repositoryML.Config, 0)
	return stg.SVC.Find(model.EntityDataRepository, &result, filters, pagination)
}

// Get returns a item
func Get(filters []stgML.Filter) (*repositoryML.Config, error) {
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
	filters := []stgML.Filter{
		{Key: model.KeyID, Value: data.ID},
	}
	data.ModifiedOn = time.Now()
	err := stg.SVC.Upsert(model.EntityDataRepository, data, filters)
	if err != nil {
		return err
	}
	busUtils.PostEvent(mcbus.TopicEventDataRepository, eventML.TypeUpdated, model.EntityDataRepository, data)
	return nil
}

// GetByID returns a item by id
func GetByID(id string) (*repositoryML.Config, error) {
	f := []stgML.Filter{
		{Key: model.KeyID, Value: id},
	}
	out := &repositoryML.Config{}
	err := stg.SVC.FindOne(model.EntityDataRepository, out, f)
	if err == nil {
		updateResult, err := updateResult(out)
		if err != nil {
			return nil, err
		}
		out = updateResult
	}
	return out, err
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []stgML.Filter{{Key: model.KeyID, Operator: stgML.OperatorIn, Value: IDs}}
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

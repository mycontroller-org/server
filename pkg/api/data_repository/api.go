package datarepository

import (
	"errors"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/json"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	repositoryML "github.com/mycontroller-org/backend/v2/pkg/model/data_repository"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]repositoryML.Config, 0)
	return stg.SVC.Find(ml.EntityDataRepository, &result, filters, pagination)
}

// Get returns a item
func Get(filters []stgml.Filter) (*repositoryML.Config, error) {
	result := &repositoryML.Config{}
	err := stg.SVC.FindOne(ml.EntityDataRepository, result, filters)
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
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: data.ID},
	}
	data.ModifiedOn = time.Now()
	return stg.SVC.Upsert(ml.EntityDataRepository, data, filters)
}

// Merge is used to update an item by task, schedule, etc
// TODO: should be updated as required
func Merge(data *repositoryML.Config) error {
	if data.ID == "" {
		return errors.New("'id' can not be empty")
	}

	// verify the supplied id is not readonly

	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: data.ID},
	}
	return stg.SVC.Upsert(ml.EntityDataRepository, data, filters)
}

// GetByID returns a item by id
func GetByID(id string) (*repositoryML.Config, error) {
	f := []stgml.Filter{
		{Key: ml.KeyID, Value: id},
	}
	out := &repositoryML.Config{}
	err := stg.SVC.FindOne(ml.EntityDataRepository, out, f)
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
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(ml.EntityDataRepository, filters)
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

package forwardpayload

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fpml "github.com/mycontroller-org/backend/v2/pkg/model/forward_payload"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]fpml.Mapping, 0)
	return stg.SVC.Find(ml.EntityForwardPayload, &result, filters, pagination)
}

// Get returns a item
func Get(filters []stgml.Filter) (*fpml.Mapping, error) {
	result := &fpml.Mapping{}
	err := stg.SVC.FindOne(ml.EntityForwardPayload, result, filters)
	return result, err
}

// Save a item details
func Save(field *fpml.Mapping) error {
	if field.ID == "" {
		field.ID = utils.RandUUID()
	}
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: field.ID},
	}
	return stg.SVC.Upsert(ml.EntityForwardPayload, field, filters)
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(ml.EntityForwardPayload, filters)
}

// Enable forward payload entries
func Enable(ids []string) error {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: ids}}
	pagination := &stgml.Pagination{Limit: 100}
	response, err := List(filters, pagination)
	if err != nil {
		return err
	}

	mappings := *response.Data.(*[]fpml.Mapping)
	for index := 0; index < len(mappings); index++ {
		mapping := mappings[index]
		if !mapping.Enabled {
			mapping.Enabled = true
			err = Save(&mapping)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Disable forward entries
func Disable(ids []string) error {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: ids}}
	pagination := &stgml.Pagination{Limit: 100}
	response, err := List(filters, pagination)
	if err != nil {
		return err
	}
	mappings := *response.Data.(*[]fpml.Mapping)
	for index := 0; index < len(mappings); index++ {
		mapping := mappings[index]
		if mapping.Enabled {
			mapping.Enabled = false
			err = Save(&mapping)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

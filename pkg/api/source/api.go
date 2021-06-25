package source

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	sourceML "github.com/mycontroller-org/server/v2/pkg/model/source"
	"github.com/mycontroller-org/server/v2/pkg/service/storage"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	storageML "github.com/mycontroller-org/server/v2/plugin/database/storage"
)

// List by filter and pagination
func List(filters []storageML.Filter, pagination *storageML.Pagination) (*storageML.Result, error) {
	result := make([]sourceML.Source, 0)
	return storage.SVC.Find(model.EntitySource, &result, filters, pagination)
}

// Get returns a source
func Get(filters []storageML.Filter) (*sourceML.Source, error) {
	result := &sourceML.Source{}
	err := storage.SVC.FindOne(model.EntitySource, result, filters)
	return result, err
}

// Save a source details
func Save(source *sourceML.Source) error {
	if source.ID == "" {
		source.ID = utils.RandUUID()
	}
	f := []storageML.Filter{
		{Key: model.KeyID, Value: source.ID},
	}
	return storage.SVC.Upsert(model.EntitySource, source, f)
}

// GetByIDs returns a source details by gatewayID, nodeId and sourceID of a message
func GetByIDs(gatewayID, nodeID, sourceID string) (*sourceML.Source, error) {
	filters := []storageML.Filter{
		{Key: model.KeyGatewayID, Value: gatewayID},
		{Key: model.KeyNodeID, Value: nodeID},
		{Key: model.KeySourceID, Value: sourceID},
	}
	result := &sourceML.Source{}
	err := storage.SVC.FindOne(model.EntitySource, result, filters)
	return result, err
}

// Delete source
func Delete(IDs []string) (int64, error) {
	filters := []storageML.Filter{{Key: model.KeyID, Operator: storageML.OperatorIn, Value: IDs}}
	return storage.SVC.Delete(model.EntitySource, filters)
}

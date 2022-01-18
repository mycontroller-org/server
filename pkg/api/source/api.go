package source

import (
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]sourceTY.Source, 0)
	return store.STORAGE.Find(types.EntitySource, &result, filters, pagination)
}

// Get returns a source
func Get(filters []storageTY.Filter) (*sourceTY.Source, error) {
	result := &sourceTY.Source{}
	err := store.STORAGE.FindOne(types.EntitySource, result, filters)
	return result, err
}

// Save a source details
func Save(source *sourceTY.Source) error {
	if source.ID == "" {
		source.ID = utils.RandUUID()
	}
	f := []storageTY.Filter{
		{Key: types.KeyID, Value: source.ID},
	}
	return store.STORAGE.Upsert(types.EntitySource, source, f)
}

// GetByIDs returns a source details by gatewayID, nodeId and sourceID of a message
func GetByIDs(gatewayID, nodeID, sourceID string) (*sourceTY.Source, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyGatewayID, Value: gatewayID},
		{Key: types.KeyNodeID, Value: nodeID},
		{Key: types.KeySourceID, Value: sourceID},
	}
	result := &sourceTY.Source{}
	err := store.STORAGE.FindOne(types.EntitySource, result, filters)
	return result, err
}

// Delete source
func Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(types.EntitySource, filters)
}

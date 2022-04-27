package source

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
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
	sources := make([]sourceTY.Source, 0)
	pagination := &storageTY.Pagination{Limit: int64(len(IDs))}
	_, err := store.STORAGE.Find(types.EntitySource, &sources, filters, pagination)
	if err != nil {
		return 0, err
	}
	deleted := int64(0)
	for _, source := range sources {
		deleteFilter := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorEqual, Value: source.ID}}
		_, err = store.STORAGE.Delete(types.EntitySource, deleteFilter)
		if err != nil {
			return deleted, err
		}
		deleted++
		// post deletion event
		busUtils.PostEvent(mcbus.TopicEventSource, eventTY.TypeDeleted, types.EntitySource, &source)
		zap.L().Info("event sent", zap.String("gatewayId", source.GatewayID), zap.String("nodeId", source.NodeID), zap.String("sourceId", source.SourceID))
	}
	return deleted, nil
}

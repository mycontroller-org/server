package source

import (
	"context"
	"fmt"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type SourceAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	bus     busTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, bus busTY.Plugin) *SourceAPI {
	return &SourceAPI{
		ctx:     ctx,
		logger:  logger.Named("source_api"),
		storage: storage,
		bus:     bus,
	}
}

// List by filter and pagination
func (s *SourceAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]sourceTY.Source, 0)
	return s.storage.Find(types.EntitySource, &result, filters, pagination)
}

// Get returns a source
func (s *SourceAPI) Get(filters []storageTY.Filter) (*sourceTY.Source, error) {
	result := &sourceTY.Source{}
	err := s.storage.FindOne(types.EntitySource, result, filters)
	return result, err
}

// Save a source details
func (s *SourceAPI) Save(source *sourceTY.Source) error {
	if source.ID == "" {
		source.ID = utils.RandUUID()
	}
	f := []storageTY.Filter{
		{Key: types.KeyID, Value: source.ID},
	}
	return s.storage.Upsert(types.EntitySource, source, f)
}

// GetByIDs returns a source details by gatewayID, nodeId and sourceID of a message
func (s *SourceAPI) GetByIDs(gatewayID, nodeID, sourceID string) (*sourceTY.Source, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyGatewayID, Value: gatewayID},
		{Key: types.KeyNodeID, Value: nodeID},
		{Key: types.KeySourceID, Value: sourceID},
	}
	result := &sourceTY.Source{}
	err := s.storage.FindOne(types.EntitySource, result, filters)
	return result, err
}

// Delete source
func (s *SourceAPI) Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	sources := make([]sourceTY.Source, 0)
	pagination := &storageTY.Pagination{Limit: int64(len(IDs))}
	_, err := s.storage.Find(types.EntitySource, &sources, filters, pagination)
	if err != nil {
		return 0, err
	}
	deleted := int64(0)
	for _, source := range sources {
		deleteFilter := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorEqual, Value: source.ID}}
		_, err = s.storage.Delete(types.EntitySource, deleteFilter)
		if err != nil {
			return deleted, err
		}
		deleted++
		// post deletion event
		busUtils.PostEvent(s.logger, s.bus, topic.TopicEventSource, eventTY.TypeDeleted, types.EntitySource, &source)
		s.logger.Debug("event sent", zap.String("gatewayId", source.GatewayID), zap.String("nodeId", source.NodeID), zap.String("sourceId", source.SourceID))
	}
	return deleted, nil
}

func (s *SourceAPI) Import(data interface{}) error {
	input, ok := data.(sourceTY.Source)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		input.ID = utils.RandUUID()
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	return s.storage.Upsert(types.EntitySource, &input, filters)
}

func (s *SourceAPI) GetEntityInterface() interface{} {
	return sourceTY.Source{}
}

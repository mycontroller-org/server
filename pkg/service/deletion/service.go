package deletion

import (
	"context"
	"fmt"

	entityAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	"github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

const (
	paginationLimit  = int64(50)
	defaultQueueSize = int(3000)
	defaultWorkers   = int(1)
)

type DeletionService struct {
	logger      *zap.Logger
	api         *entityAPI.API
	bus         busTY.Plugin
	eventsQueue *queueUtils.QueueSpec
}

func New(ctx context.Context) (serviceTY.Service, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	api, err := entityAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	svc := &DeletionService{
		logger: logger.Named("resource_deletion_service"),
		api:    api,
		bus:    bus,
	}
	svc.eventsQueue = &queueUtils.QueueSpec{
		Topic:          topic.TopicEventsAll,
		Queue:          queueUtils.New(svc.logger, "deletion_service", defaultQueueSize, svc.processEvent, defaultWorkers),
		SubscriptionId: -1,
	}

	return svc, nil
}

func (svc *DeletionService) Name() string {
	return "deletion_service"
}

// Start event process engine
func (svc *DeletionService) Start() error {
	// add received events in to local queue
	sID, err := svc.bus.Subscribe(svc.eventsQueue.Topic, svc.onEventReceive)
	if err != nil {
		return err
	}
	svc.eventsQueue.SubscriptionId = sID
	return nil
}

func (svc *DeletionService) Close() error {
	err := svc.bus.Unsubscribe(svc.eventsQueue.Topic, svc.eventsQueue.SubscriptionId)
	if err != nil {
		return err
	}
	svc.eventsQueue.Close()
	return nil
}

func (svc *DeletionService) onEventReceive(busData *busTY.BusData) {
	status := svc.eventsQueue.Produce(busData)
	if !status {
		svc.logger.Warn("failed to store the event into queue", zap.Any("event", busData))
	}
}

func (svc *DeletionService) processEvent(item interface{}) {
	busData := item.(*busTY.BusData)
	event := &eventTY.Event{}
	err := busData.LoadData(event)
	if err != nil {
		svc.logger.Warn("error on convert to target type", zap.Any("topic", busData.Topic), zap.Error(err))
		return
	}

	// if it is not a deletion event, return from here
	if event.Type != eventTY.TypeDeleted {
		return
	}

	svc.logger.Debug("received an deletion event", zap.Any("event", event))

	// supported entity events
	switch event.EntityType {

	case types.EntityGateway:
		gateway := &gatewayTY.Config{}
		err = event.LoadEntity(gateway)
		if err != nil {
			svc.logger.Warn("error on loading entity", zap.Any("event", event), zap.Error(err))
			return
		}
		svc.deleteNodes(gateway)

	case types.EntityNode:
		node := &nodeTY.Node{}
		err = event.LoadEntity(node)
		if err != nil {
			svc.logger.Warn("error on loading entity", zap.Any("event", event), zap.Error(err))
			return
		}
		svc.deleteSources(node)

	case types.EntitySource:
		source := &sourceTY.Source{}
		err = event.LoadEntity(source)
		if err != nil {
			svc.logger.Warn("error on loading entity", zap.Any("event", event), zap.Error(err))
			return
		}
		svc.deleteFields(source)

	default:
		// do not proceed further
		return
	}
}

// deletes nodes
func (svc *DeletionService) deleteNodes(gateway *gatewayTY.Config) {
	filters := []storageTY.Filter{{Key: types.KeyGatewayID, Operator: storageTY.OperatorEqual, Value: gateway.ID}}
	pagination := &storageTY.Pagination{Limit: paginationLimit, Offset: 0}
	for {
		result, err := svc.api.Node().List(filters, pagination)
		if err != nil {
			svc.logger.Error("error on getting nodes list", zap.String("gatewayId", gateway.ID), zap.Int64("offset", pagination.Offset), zap.Error(err))
			return
		}

		if result.Count == 0 {
			break
		}

		// collect node ids and delete those
		nodes, ok := result.Data.(*[]nodeTY.Node)
		if !ok {
			svc.logger.Error("error on casting to nodes", zap.String("originalType", fmt.Sprintf("%T", result.Data)))
			return
		}
		nodeIDs := []string{}
		for _, node := range *nodes {
			nodeIDs = append(nodeIDs, node.ID)
		}
		_, err = svc.api.Node().Delete(nodeIDs)
		if err != nil {
			svc.logger.Error("error on deleting nodes", zap.Any("nodeIds", nodeIDs), zap.Error(err))
			return
		}

		// if page over terminate from the loop
		if result.Count < paginationLimit {
			break
		}
	}
}

// deletes sources
func (svc *DeletionService) deleteSources(node *nodeTY.Node) {
	filters := []storageTY.Filter{
		{Key: types.KeyGatewayID, Operator: storageTY.OperatorEqual, Value: node.GatewayID},
		{Key: types.KeyNodeID, Operator: storageTY.OperatorEqual, Value: node.NodeID},
	}
	pagination := &storageTY.Pagination{Limit: paginationLimit, Offset: 0}
	for {
		result, err := svc.api.Source().List(filters, pagination)
		if err != nil {
			svc.logger.Error("error on getting sources list", zap.String("gatewayId", node.GatewayID), zap.String("nodeId", node.NodeID), zap.Int64("offset", pagination.Offset), zap.Error(err))
			return
		}

		if result.Count == 0 {
			break
		}

		// collect source ids and delete those
		sources, ok := result.Data.(*[]sourceTY.Source)
		if !ok {
			svc.logger.Error("error on casting to sources", zap.String("originalType", fmt.Sprintf("%T", result.Data)))
			return
		}
		sourceIDs := []string{}
		for _, source := range *sources {
			sourceIDs = append(sourceIDs, source.ID)
		}
		_, err = svc.api.Source().Delete(sourceIDs)
		if err != nil {
			svc.logger.Error("error on deleting sources", zap.Any("sourceIds", sourceIDs), zap.Error(err))
			return
		}

		// if page over terminate from the loop
		if result.Count < paginationLimit {
			break
		}
	}
}

// deletes fields
func (svc *DeletionService) deleteFields(source *sourceTY.Source) {
	filters := []storageTY.Filter{
		{Key: types.KeyGatewayID, Operator: storageTY.OperatorEqual, Value: source.GatewayID},
		{Key: types.KeyNodeID, Operator: storageTY.OperatorEqual, Value: source.NodeID},
		{Key: types.KeySourceID, Operator: storageTY.OperatorEqual, Value: source.SourceID},
	}
	pagination := &storageTY.Pagination{Limit: paginationLimit, Offset: 0}
	for {
		result, err := svc.api.Field().List(filters, pagination)
		if err != nil {
			svc.logger.Error("error on getting sources list", zap.String("gatewayId", source.GatewayID), zap.String("nodeId", source.NodeID), zap.String("sourceId", source.SourceID), zap.Int64("offset", pagination.Offset), zap.Error(err))
			return
		}

		if result.Count == 0 {
			svc.logger.Info("no records found")
			break
		}

		// collect field ids and delete those
		fields, ok := result.Data.(*[]fieldTY.Field)
		if !ok {
			svc.logger.Error("error on casting to fields", zap.String("originalType", fmt.Sprintf("%T", result.Data)))
			return
		}
		fieldIDs := []string{}
		for _, field := range *fields {
			fieldIDs = append(fieldIDs, field.ID)
		}
		_, err = svc.api.Field().Delete(fieldIDs)
		if err != nil {
			svc.logger.Error("error on deleting fields", zap.Any("fieldIds", fieldIDs), zap.Error(err))
			return
		}

		// if page over terminate from the loop
		if result.Count < paginationLimit {
			break
		}
	}
}

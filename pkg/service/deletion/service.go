package deletion

import (
	"fmt"

	fieldAPI "github.com/mycontroller-org/server/v2/pkg/api/field"
	nodeAPI "github.com/mycontroller-org/server/v2/pkg/api/node"
	sourceAPI "github.com/mycontroller-org/server/v2/pkg/api/source"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/types"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

var (
	eventsQueue          *queueUtils.Queue
	queueSize            = int(3000)
	workers              = 1
	eventsTopic          = ""
	eventsSubscriptionID = int64(0)
)

const (
	paginationLimit = int64(50)
)

// Start event process engine
func Start() error {
	eventsQueue = queueUtils.New("deletion_service", queueSize, processEvent, workers)

	// add received events in to local queue
	eventsTopic = mcbus.FormatTopic(mcbus.TopicEventsAll)
	sID, err := mcbus.Subscribe(eventsTopic, onEventReceive)
	if err != nil {
		return err
	}
	eventsSubscriptionID = sID
	return nil
}

func Close() error {
	err := mcbus.Unsubscribe(eventsTopic, eventsSubscriptionID)
	if err != nil {
		return err
	}
	eventsQueue.Close()
	return nil
}

func onEventReceive(busData *busTY.BusData) {
	status := eventsQueue.Produce(busData)
	if !status {
		zap.L().Warn("failed to store the event into queue", zap.Any("event", busData))
	}
}

func processEvent(item interface{}) {
	busData := item.(*busTY.BusData)
	event := &eventTY.Event{}
	err := busData.LoadData(event)
	if err != nil {
		zap.L().Warn("error on convet to target type", zap.Any("topic", busData.Topic), zap.Error(err))
		return
	}

	// if it is not a deletion event, return from here
	if event.Type != eventTY.TypeDeleted {
		return
	}

	zap.L().Debug("received an deleton event", zap.Any("event", event))

	// supported entity events
	switch event.EntityType {

	case types.EntityGateway:
		gateway := &gatewayTY.Config{}
		err = event.LoadEntity(gateway)
		if err != nil {
			zap.L().Warn("error on loading entity", zap.Any("event", event), zap.Error(err))
			return
		}
		deleteNodes(gateway)

	case types.EntityNode:
		node := &nodeTY.Node{}
		err = event.LoadEntity(node)
		if err != nil {
			zap.L().Warn("error on loading entity", zap.Any("event", event), zap.Error(err))
			return
		}
		deleteSources(node)

	case types.EntitySource:
		source := &sourceTY.Source{}
		err = event.LoadEntity(source)
		if err != nil {
			zap.L().Warn("error on loading entity", zap.Any("event", event), zap.Error(err))
			return
		}
		deleteFields(source)

	default:
		// do not proceed further
		return
	}
}

// deletes nodes
func deleteNodes(gateway *gatewayTY.Config) {
	filters := []storageTY.Filter{{Key: types.KeyGatewayID, Operator: storageTY.OperatorEqual, Value: gateway.ID}}
	pagination := &storageTY.Pagination{Limit: paginationLimit, Offset: 0}
	for {
		result, err := nodeAPI.List(filters, pagination)
		if err != nil {
			zap.L().Error("error on geting nodes list", zap.String("gatewayId", gateway.ID), zap.Int64("offset", pagination.Offset), zap.Error(err))
			return
		}

		if result.Count == 0 {
			break
		}

		// collect node ids and delete those
		nodes, ok := result.Data.(*[]nodeTY.Node)
		if !ok {
			zap.L().Error("error on casting to nodes", zap.String("originalType", fmt.Sprintf("%T", result.Data)))
			return
		}
		nodeIDs := []string{}
		for _, node := range *nodes {
			nodeIDs = append(nodeIDs, node.ID)
		}
		_, err = nodeAPI.Delete(nodeIDs)
		if err != nil {
			zap.L().Error("error on deleting nodes", zap.Any("nodeIds", nodeIDs), zap.Error(err))
			return
		}

		// if page over terminate from the loop
		if result.Count < paginationLimit {
			break
		}
	}
}

// deletes sources
func deleteSources(node *nodeTY.Node) {
	filters := []storageTY.Filter{
		{Key: types.KeyGatewayID, Operator: storageTY.OperatorEqual, Value: node.GatewayID},
		{Key: types.KeyNodeID, Operator: storageTY.OperatorEqual, Value: node.NodeID},
	}
	pagination := &storageTY.Pagination{Limit: paginationLimit, Offset: 0}
	for {
		result, err := sourceAPI.List(filters, pagination)
		if err != nil {
			zap.L().Error("error on geting sources list", zap.String("gatewayId", node.GatewayID), zap.String("nodeId", node.NodeID), zap.Int64("offset", pagination.Offset), zap.Error(err))
			return
		}

		if result.Count == 0 {
			break
		}

		// collect source ids and delete those
		sources, ok := result.Data.(*[]sourceTY.Source)
		if !ok {
			zap.L().Error("error on casting to sources", zap.String("originalType", fmt.Sprintf("%T", result.Data)))
			return
		}
		sourceIDs := []string{}
		for _, source := range *sources {
			sourceIDs = append(sourceIDs, source.ID)
		}
		_, err = sourceAPI.Delete(sourceIDs)
		if err != nil {
			zap.L().Error("error on deleting sources", zap.Any("sourceIds", sourceIDs), zap.Error(err))
			return
		}

		// if page over terminate from the loop
		if result.Count < paginationLimit {
			break
		}
	}
}

// deletes fields
func deleteFields(source *sourceTY.Source) {
	filters := []storageTY.Filter{
		{Key: types.KeyGatewayID, Operator: storageTY.OperatorEqual, Value: source.GatewayID},
		{Key: types.KeyNodeID, Operator: storageTY.OperatorEqual, Value: source.NodeID},
		{Key: types.KeySourceID, Operator: storageTY.OperatorEqual, Value: source.SourceID},
	}
	pagination := &storageTY.Pagination{Limit: paginationLimit, Offset: 0}
	for {
		result, err := fieldAPI.List(filters, pagination)
		if err != nil {
			zap.L().Error("error on geting sources list", zap.String("gatewayId", source.GatewayID), zap.String("nodeId", source.NodeID), zap.String("sourceId", source.SourceID), zap.Int64("offset", pagination.Offset), zap.Error(err))
			return
		}

		if result.Count == 0 {
			zap.L().Info("no records found")
			break
		}

		// collect field ids and delete those
		fields, ok := result.Data.(*[]fieldTY.Field)
		if !ok {
			zap.L().Error("error on casting to fields", zap.String("originalType", fmt.Sprintf("%T", result.Data)))
			return
		}
		fieldIDs := []string{}
		for _, field := range *fields {
			fieldIDs = append(fieldIDs, field.ID)
		}
		_, err = fieldAPI.Delete(fieldIDs)
		if err != nil {
			zap.L().Error("error on deleting fields", zap.Any("fieldIds", fieldIDs), zap.Error(err))
			return
		}

		// if page over terminate from the loop
		if result.Count < paginationLimit {
			break
		}
	}
}

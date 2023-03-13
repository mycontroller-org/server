package node

import (
	"context"
	"errors"
	"fmt"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type NodeAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
	bus     busTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin, bus busTY.Plugin) *NodeAPI {
	return &NodeAPI{
		ctx:     ctx,
		logger:  logger.Named("node_api"),
		storage: storage,
		bus:     bus,
	}
}

// List by filter and pagination
func (n *NodeAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]nodeTY.Node, 0)
	return n.storage.Find(types.EntityNode, &result, filters, pagination)
}

// Get returns a Node
func (n *NodeAPI) Get(filters []storageTY.Filter) (nodeTY.Node, error) {
	result := nodeTY.Node{}
	err := n.storage.FindOne(types.EntityNode, &result, filters)
	return result, err
}

// Save Node config into disk
func (n *NodeAPI) Save(node *nodeTY.Node, publishEvent bool) error {
	eventType := eventTY.TypeUpdated
	if node.ID == "" {
		node.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: node.ID},
	}
	err := n.storage.Upsert(types.EntityNode, node, filters)
	if err != nil {
		return err
	}
	if publishEvent {
		// post node data to event listeners
		busUtils.PostEvent(n.logger, n.bus, topic.TopicEventNode, eventType, types.EntityNode, node)
	}
	return nil
}

// GetByGatewayAndNodeID returns a node details by gatewayID and nodeId of a message
func (n *NodeAPI) GetByGatewayAndNodeID(gatewayID, nodeID string) (*nodeTY.Node, error) {
	f := []storageTY.Filter{
		{Key: types.KeyGatewayID, Value: gatewayID},
		{Key: types.KeyNodeID, Value: nodeID},
	}
	node := nodeTY.Node{}
	err := n.storage.FindOne(types.EntityNode, &node, f)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetByID returns a node details by id
func (n *NodeAPI) GetByID(id string) (*nodeTY.Node, error) {
	f := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	result := &nodeTY.Node{}
	err := n.storage.FindOne(types.EntityNode, result, f)
	return result, err
}

// GetByIDs returns a node details by id
func (n *NodeAPI) GetByIDs(ids []string) ([]nodeTY.Node, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids},
	}
	pagination := &storageTY.Pagination{Limit: int64(len(ids))}
	nodes := make([]nodeTY.Node, 0)
	_, err := n.storage.Find(types.EntityNode, &nodes, filters, pagination)
	return nodes, err
}

// Delete node
func (n *NodeAPI) Delete(IDs []string) (int64, error) {
	nodes, err := n.GetByIDs(IDs)
	if err != nil {
		return 0, err
	}

	// delete one by one and report deletion event
	deleted := int64(0)
	for _, node := range nodes {
		filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorEqual, Value: node.ID}}
		_, err = n.storage.Delete(types.EntityNode, filters)
		if err != nil {
			return deleted, err
		}
		deleted++
		busUtils.PostEvent(n.logger, n.bus, topic.TopicEventNode, eventTY.TypeDeleted, types.EntityNode, node)
	}

	return deleted, nil
}

// UpdateFirmwareState func
func (n *NodeAPI) UpdateFirmwareState(id string, data map[string]interface{}) error {
	if id == "" {
		return errors.New("id not supplied")
	}
	node, err := n.GetByID(id)
	if err != nil {
		return err
	}

	// update fields
	node.Others.Set(types.FieldOTARunning, utils.GetMapValue(data, types.FieldOTARunning, nil), nil)
	node.Others.Set(types.FieldOTABlockNumber, utils.GetMapValue(data, types.FieldOTABlockNumber, nil), nil)
	node.Others.Set(types.FieldOTAProgress, utils.GetMapValue(data, types.FieldOTAProgress, nil), nil)
	node.Others.Set(types.FieldOTAStatusOn, utils.GetMapValue(data, types.FieldOTAStatusOn, nil), nil)
	node.Others.Set(types.FieldOTABlockTotal, utils.GetMapValue(data, types.FieldOTABlockTotal, nil), nil)

	// start time
	startTime := utils.GetMapValue(data, types.FieldOTAStartTime, nil)
	if startTime != nil {
		node.Others.Set(types.FieldOTAStartTime, startTime, nil)
		node.Others.Set(types.FieldOTATimeTaken, "", nil)
		node.Others.Set(types.FieldOTAEndTime, "", nil)
	}

	endTime := utils.GetMapValue(data, types.FieldOTAEndTime, nil)
	if endTime != nil {
		node.Others.Set(types.FieldOTAEndTime, endTime, nil)
		startTime = node.Others.Get(types.FieldOTAStartTime)
		if st, stOK := startTime.(time.Time); stOK {
			if et, etOK := endTime.(time.Time); etOK {
				node.Others.Set(types.FieldOTATimeTaken, et.Sub(st).String(), nil)
			}
		}
	}

	return n.Save(node, true)
}

// Verifies node up status by checking the last seen timestamp
// if the last seen greater than x minutes/seconds or specified duration in that node
// will be marked as down
func (n *NodeAPI) VerifyNodeUpStatus(inactiveDuration time.Duration) {
	filters := []storageTY.Filter{{Key: "State.Status", Value: types.StatusUp, Operator: storageTY.OperatorEqual}}
	limit := int64(50)
	offset := int64(0)
	pagination := &storageTY.Pagination{
		Limit:  limit,
		Offset: offset,
		SortBy: []storageTY.Sort{{Field: types.KeyID, OrderBy: storageTY.SortByASC}},
	}

	for {
		result, err := n.List(filters, pagination)
		if err != nil {
			n.logger.Error("error on getting active nodes list", zap.Error(err))
			return
		}
		if result.Count == 0 {
			return
		}

		// process received nodes
		n.updateNodesUpStatus(result, inactiveDuration)

		if result.Count < limit {
			return
		}
		// move to next page
		offset++
	}
}

func (n *NodeAPI) updateNodesUpStatus(result *storageTY.Result, inactiveDuration time.Duration) {
	nodesPointer, ok := result.Data.(*[]nodeTY.Node)
	if !ok {
		n.logger.Error("invalid data", zap.String("receivedType", fmt.Sprintf("%T", result.Data)))
		return
	}
	nodes := *nodesPointer

	// lastSeen marker
	currentTime := time.Now()
	for index := range nodes {
		node := nodes[index]
		// get custom inactive reference
		strDuration := node.Labels.Get(types.LabelNodeInactiveDuration)
		duration := utils.ToDuration(strDuration, inactiveDuration)
		inactiveReference := currentTime.Add(-duration)
		if node.State.Status == types.StatusUp && node.LastSeen.Before(inactiveReference) {
			node.State = types.State{
				Status:  types.StatusDown,
				Since:   currentTime,
				Message: "marked by server",
			}
			err := n.Save(&node, true)
			if err != nil {
				n.logger.Error("error on saving a node status", zap.String("gatewayId", node.GatewayID), zap.String("nodeId", node.NodeID), zap.Error(err))
			}
		}
	}
}

func (n *NodeAPI) Import(data interface{}) error {
	input, ok := data.(nodeTY.Node)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		input.ID = utils.RandUUID()
	}

	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}
	return n.storage.Upsert(types.EntityNode, &input, filters)
}

func (n *NodeAPI) GetEntityInterface() interface{} {
	return nodeTY.Node{}
}

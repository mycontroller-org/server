package node

import (
	"errors"
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]nodeTY.Node, 0)
	return store.STORAGE.Find(types.EntityNode, &result, filters, pagination)
}

// Get returns a Node
func Get(filters []storageTY.Filter) (nodeTY.Node, error) {
	result := nodeTY.Node{}
	err := store.STORAGE.FindOne(types.EntityNode, &result, filters)
	return result, err
}

// Save Node config into disk
func Save(node *nodeTY.Node, publishEvent bool) error {
	eventType := eventTY.TypeUpdated
	if node.ID == "" {
		node.ID = utils.RandUUID()
		eventType = eventTY.TypeCreated
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: node.ID},
	}
	err := store.STORAGE.Upsert(types.EntityNode, node, filters)
	if err != nil {
		return err
	}
	if publishEvent {
		// post node data to event listeners
		busUtils.PostEvent(mcbus.TopicEventNode, eventType, types.EntityNode, node)
	}
	return nil
}

// GetByGatewayAndNodeID returns a node details by gatewayID and nodeId of a message
func GetByGatewayAndNodeID(gatewayID, nodeID string) (*nodeTY.Node, error) {
	f := []storageTY.Filter{
		{Key: types.KeyGatewayID, Value: gatewayID},
		{Key: types.KeyNodeID, Value: nodeID},
	}
	node := nodeTY.Node{}
	err := store.STORAGE.FindOne(types.EntityNode, &node, f)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetByID returns a node details by id
func GetByID(id string) (*nodeTY.Node, error) {
	f := []storageTY.Filter{
		{Key: types.KeyID, Value: id},
	}
	result := &nodeTY.Node{}
	err := store.STORAGE.FindOne(types.EntityNode, result, f)
	return result, err
}

// GetByIDs returns a node details by id
func GetByIDs(ids []string) ([]nodeTY.Node, error) {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: ids},
	}
	pagination := &storageTY.Pagination{Limit: int64(len(ids))}
	nodes := make([]nodeTY.Node, 0)
	_, err := store.STORAGE.Find(types.EntityNode, &nodes, filters, pagination)
	return nodes, err
}

// Delete node
func Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(types.EntityNode, filters)
}

// UpdateFirmwareState func
func UpdateFirmwareState(id string, data map[string]interface{}) error {
	if id == "" {
		return errors.New("id not supplied")
	}
	node, err := GetByID(id)
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

	return Save(node, true)
}

// Verifies node up status by checking the last seen timestamp
// if the last seen greater than x minutes/seconds or specified duration in that node
// will be marked as down
func VerifyNodeUpStatus(inactiveDuration time.Duration) {
	filters := []storageTY.Filter{{Key: "State.Status", Value: types.StatusUp, Operator: storageTY.OperatorEqual}}
	limit := int64(50)
	offset := int64(0)
	pagination := &storageTY.Pagination{
		Limit:  limit,
		Offset: offset,
		SortBy: []storageTY.Sort{{Field: types.KeyID, OrderBy: storageTY.SortByASC}},
	}

	for {
		result, err := List(filters, pagination)
		if err != nil {
			zap.L().Error("error on getting active nodes list", zap.Error(err))
			return
		}
		if result.Count == 0 {
			return
		}

		// process received nodes
		updateNodesUpStatus(result, inactiveDuration)

		if result.Count < limit {
			return
		}
		// move to next page
		offset++
	}
}

func updateNodesUpStatus(result *storageTY.Result, inactiveDuration time.Duration) {
	nodesPointer, ok := result.Data.(*[]nodeTY.Node)
	if !ok {
		zap.L().Error("invalid data", zap.String("receivedType", fmt.Sprintf("%T", result.Data)))
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
			err := Save(&node, true)
			if err != nil {
				zap.L().Error("error on saving a node status", zap.String("gatewayId", node.GatewayID), zap.String("nodeId", node.NodeID), zap.Error(err))
			}
		}
	}
}

package node

import (
	"errors"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
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
func Save(node *nodeTY.Node) error {
	if node.ID == "" {
		node.ID = utils.RandUUID()
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: node.ID},
	}
	return store.STORAGE.Upsert(types.EntityNode, node, filters)
}

// GetByGatewayAndNodeID returns a node details by gatewayID and nodeId of a message
func GetByGatewayAndNodeID(gatewayID, nodeID string) (*nodeTY.Node, error) {
	f := []storageTY.Filter{
		{Key: types.KeyGatewayID, Value: gatewayID},
		{Key: types.KeyNodeID, Value: nodeID},
	}
	result := &nodeTY.Node{}
	err := store.STORAGE.FindOne(types.EntityNode, result, f)
	return result, err
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

	return Save(node)
}

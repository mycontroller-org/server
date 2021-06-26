package node

import (
	"errors"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/model"
	nodeML "github.com/mycontroller-org/server/v2/pkg/model/node"
	stg "github.com/mycontroller-org/server/v2/pkg/service/database/storage"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	stgML "github.com/mycontroller-org/server/v2/plugin/database/storage"
)

// List by filter and pagination
func List(filters []stgML.Filter, pagination *stgML.Pagination) (*stgML.Result, error) {
	result := make([]nodeML.Node, 0)
	return stg.SVC.Find(model.EntityNode, &result, filters, pagination)
}

// Get returns a Node
func Get(filters []stgML.Filter) (nodeML.Node, error) {
	result := nodeML.Node{}
	err := stg.SVC.FindOne(model.EntityNode, &result, filters)
	return result, err
}

// Save Node config into disk
func Save(node *nodeML.Node) error {
	if node.ID == "" {
		node.ID = utils.RandUUID()
	}
	filters := []stgML.Filter{
		{Key: model.KeyID, Value: node.ID},
	}
	return stg.SVC.Upsert(model.EntityNode, node, filters)
}

// GetByGatewayAndNodeID returns a node details by gatewayID and nodeId of a message
func GetByGatewayAndNodeID(gatewayID, nodeID string) (*nodeML.Node, error) {
	f := []stgML.Filter{
		{Key: model.KeyGatewayID, Value: gatewayID},
		{Key: model.KeyNodeID, Value: nodeID},
	}
	result := &nodeML.Node{}
	err := stg.SVC.FindOne(model.EntityNode, result, f)
	return result, err
}

// GetByID returns a node details by id
func GetByID(id string) (*nodeML.Node, error) {
	f := []stgML.Filter{
		{Key: model.KeyID, Value: id},
	}
	result := &nodeML.Node{}
	err := stg.SVC.FindOne(model.EntityNode, result, f)
	return result, err
}

// GetByIDs returns a node details by id
func GetByIDs(ids []string) ([]nodeML.Node, error) {
	filters := []stgML.Filter{
		{Key: model.KeyID, Operator: stgML.OperatorIn, Value: ids},
	}
	pagination := &stgML.Pagination{Limit: int64(len(ids))}
	nodes := make([]nodeML.Node, 0)
	_, err := stg.SVC.Find(model.EntityNode, &nodes, filters, pagination)
	return nodes, err
}

// Delete node
func Delete(IDs []string) (int64, error) {
	filters := []stgML.Filter{{Key: model.KeyID, Operator: stgML.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(model.EntityNode, filters)
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
	node.Others.Set(model.FieldOTARunning, utils.GetMapValue(data, model.FieldOTARunning, nil), nil)
	node.Others.Set(model.FieldOTABlockNumber, utils.GetMapValue(data, model.FieldOTABlockNumber, nil), nil)
	node.Others.Set(model.FieldOTAProgress, utils.GetMapValue(data, model.FieldOTAProgress, nil), nil)
	node.Others.Set(model.FieldOTAStatusOn, utils.GetMapValue(data, model.FieldOTAStatusOn, nil), nil)
	node.Others.Set(model.FieldOTABlockTotal, utils.GetMapValue(data, model.FieldOTABlockTotal, nil), nil)

	// start time
	startTime := utils.GetMapValue(data, model.FieldOTAStartTime, nil)
	if startTime != nil {
		node.Others.Set(model.FieldOTAStartTime, startTime, nil)
		node.Others.Set(model.FieldOTATimeTaken, "", nil)
		node.Others.Set(model.FieldOTAEndTime, "", nil)
	}

	endTime := utils.GetMapValue(data, model.FieldOTAEndTime, nil)
	if endTime != nil {
		node.Others.Set(model.FieldOTAEndTime, endTime, nil)
		startTime = node.Others.Get(model.FieldOTAStartTime)
		if st, stOK := startTime.(time.Time); stOK {
			if et, etOK := endTime.(time.Time); etOK {
				node.Others.Set(model.FieldOTATimeTaken, et.Sub(st).String(), nil)
			}
		}
	}

	return Save(node)
}

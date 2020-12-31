package node

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]nml.Node, 0)
	return stg.SVC.Find(ml.EntityNode, &result, filters, pagination)
}

// Get returns a Node
func Get(filters []stgml.Filter) (nml.Node, error) {
	result := nml.Node{}
	err := stg.SVC.FindOne(ml.EntityNode, &result, filters)
	return result, err
}

// Save Node config into disk
func Save(node *nml.Node) error {
	if node.ID == "" {
		node.ID = ut.RandUUID()
	}
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: node.ID},
	}
	return stg.SVC.Upsert(ml.EntityNode, node, filters)
}

// GetByIDs returns a node details by gatewayID and nodeId of a message
func GetByIDs(gatewayID, nodeID string) (*nml.Node, error) {
	f := []stgml.Filter{
		{Key: ml.KeyGatewayID, Value: gatewayID},
		{Key: ml.KeyNodeID, Value: nodeID},
	}
	result := &nml.Node{}
	err := stg.SVC.FindOne(ml.EntityNode, result, f)
	return result, err
}

// Delete node
func Delete(IDs []string) (int64, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(ml.EntityNode, filters)
}

package node

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	stgml "github.com/mycontroller-org/backend/v2/pkg/model/storage"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
)

// List by filter and pagination
func List(filters []pml.Filter, pagination *pml.Pagination) (*pml.Result, error) {
	result := make([]nml.Node, 0)
	return svc.STG.Find(ml.EntityNode, &result, filters, pagination)
}

// Get returns a Node
func Get(filters []pml.Filter) (nml.Node, error) {
	result := nml.Node{}
	err := svc.STG.FindOne(ml.EntityNode, &result, filters)
	return result, err
}

// Save Node config into disk
func Save(node *nml.Node) error {
	if node.ID == "" {
		node.ID = ut.RandUUID()
	}
	filters := []pml.Filter{
		{Key: ml.KeyID, Value: node.ID},
	}
	return svc.STG.Upsert(ml.EntityNode, node, filters)
}

// GetByIDs returns a node details by gatewayID and nodeId of a message
func GetByIDs(gatewayID, nodeID string) (*nml.Node, error) {
	f := []pml.Filter{
		{Key: ml.KeyGatewayID, Value: gatewayID},
		{Key: ml.KeyNodeID, Value: nodeID},
	}
	result := &nml.Node{}
	err := svc.STG.FindOne(ml.EntityNode, result, f)
	return result, err
}

// Delete node
func Delete(IDs []string) (int64, error) {
	filters := []pml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return svc.STG.Delete(ml.EntityNode, filters)
}

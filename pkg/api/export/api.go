package export

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(f []stgml.Filter, p *stgml.Pagination) (*stgml.Result, error) {
	out := make([]nml.Node, 0)
	return stg.SVC.Find(ml.EntityNode, &out, f, p)
}

// Get returns a Node
func Get(f []stgml.Filter) (nml.Node, error) {
	out := nml.Node{}
	err := stg.SVC.FindOne(ml.EntityNode, &out, f)
	return out, err
}

// Save Node config into disk
func Save(node *nml.Node) error {
	if node.ID == "" {
		node.ID = ut.RandUUID()
	}
	f := []stgml.Filter{
		{Key: ml.KeyID, Value: node.ID},
	}
	return stg.SVC.Upsert(ml.EntityNode, node, f)
}

// GetByIDs returns a node details by gatewayID and nodeId of a message
func GetByIDs(gatewayID, nodeID string) (*nml.Node, error) {
	f := []stgml.Filter{
		{Key: ml.KeyGatewayID, Value: gatewayID},
		{Key: ml.KeyNodeID, Value: nodeID},
	}
	out := &nml.Node{}
	err := stg.SVC.FindOne(ml.EntityNode, out, f)
	return out, err
}

// Delete node
func Delete(IDs []string) (int64, error) {
	f := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(ml.EntityNode, f)
}

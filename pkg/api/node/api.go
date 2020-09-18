package node

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	"github.com/mycontroller-org/backend/v2/pkg/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
)

// List by filter and pagination
func List(f []pml.Filter, p *pml.Pagination) ([]nml.Node, error) {
	out := make([]nml.Node, 0)
	svc.STG.Find(ml.EntityNode, f, p, &out)
	return out, nil
}

// Get returns a Node
func Get(f []pml.Filter) (nml.Node, error) {
	out := nml.Node{}
	err := svc.STG.FindOne(ml.EntityNode, f, &out)
	return out, err
}

// Save Node config into disk
func Save(node *nml.Node) error {
	if node.ID == "" {
		node.ID = ut.RandUUID()
	}
	f := []pml.Filter{
		{Key: ml.KeyID, Value: node.ID},
	}
	return svc.STG.Upsert(ml.EntityNode, f, node)
}

// GetByIDs returns a node details by gatewayID and nodeId of a message
func GetByIDs(gatewayID, nodeID string) (*nml.Node, error) {
	f := []pml.Filter{
		{Key: ml.KeyGatewayID, Value: gatewayID},
		{Key: ml.KeyNodeID, Value: nodeID},
	}
	out := &nml.Node{}
	err := svc.STG.FindOne(ml.EntityNode, f, out)
	return out, err
}

// Delete node
func Delete(IDs []string) (int64, error) {
	f := []pml.Filter{{Key: ml.KeyID, Operator: storage.OperatorIn, Value: IDs}}
	return svc.STG.Delete(ml.EntityNode, f)
}

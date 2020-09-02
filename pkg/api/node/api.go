package node

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
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
		node.ID = ut.RandID()
	}
	f := []pml.Filter{
		{Key: "id", Operator: "eq", Value: node.ID},
	}
	return svc.STG.Upsert(ml.EntityNode, f, node)
}

// GetByIDs returns a node details by gatewayID and nodeId of a message
func GetByIDs(gatewayID, nodeID string) (*nml.Node, error) {
	id := nml.AssembleID(gatewayID, nodeID)
	f := []pml.Filter{
		{Key: "id", Operator: "eq", Value: id},
	}
	out := &nml.Node{}
	err := svc.STG.FindOne(ml.EntityNode, f, out)
	return out, err
}

package node

import (
	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	svc "github.com/mycontroller-org/mycontroller-v2/pkg/service"
	ut "github.com/mycontroller-org/mycontroller-v2/pkg/util"
)

// ListNodes by filter and pagination
func ListNodes(f []ml.Filter, p ml.Pagination) ([]ml.Node, error) {
	out := make([]ml.Node, 0)
	svc.STG.Find(ml.EntityNode, f, p, &out)
	return out, nil
}

// GetNode returns a Node
func GetNode(f []ml.Filter) (ml.Node, error) {
	out := ml.Node{}
	err := svc.STG.FindOne(ml.EntityNode, f, &out)
	return out, err
}

// Save Node config into disk
func Save(g *ml.Node) error {
	if g.ID == "" {
		g.ID = ut.RandID()
	}
	return svc.STG.Upsert(ml.EntityNode, nil, g)
}

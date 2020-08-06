package gateway

import (
	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	svc "github.com/mycontroller-org/mycontroller-v2/pkg/service"
	ut "github.com/mycontroller-org/mycontroller-v2/pkg/util"
)

// ListGateways by filter and pagination
func ListGateways(f []ml.Filter, p ml.Pagination) ([]ml.GatewayConfig, error) {
	out := make([]ml.GatewayConfig, 0)
	svc.STG.Find(ml.EntityGateway, f, p, &out)
	return out, nil
}

// GetGateway returns a gateway
func GetGateway(f []ml.Filter) (ml.GatewayConfig, error) {
	out := ml.GatewayConfig{}
	err := svc.STG.FindOne(ml.EntityGateway, f, &out)
	return out, err
}

// Save gateway config into disk
func Save(g *ml.GatewayConfig) error {
	if g.ID == "" {
		g.ID = ut.RandID()
	}
	return svc.STG.Upsert(ml.EntityGateway, nil, g)
}

// SetState Updates state data
func SetState(g *ml.GatewayConfig, s ml.State) error {
	g.State = s
	return Save(g)
}

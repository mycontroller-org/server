package sample

import (
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gw_protocol"
)

// Provider implementation
type Provider struct {
	GWConfig *gwml.Config
	Gateway  gwpl.Gateway
}

// Post func
func (p *Provider) Post(rawMsg *msgml.RawMessage) error {
	return p.Gateway.Write(rawMsg)
}

// Start func
func (p *Provider) Start(rxMessageFunc func(rawMsg *msgml.RawMessage) error) error {
	return nil
}

// Close func
func (p *Provider) Close() error {
	// do internal works
	// close gateway
	return p.Gateway.Close()
}

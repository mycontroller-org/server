package philipshue

import (
	"fmt"

	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gw_protocol"
	httpProtocol "github.com/mycontroller-org/backend/v2/plugin/gw_protocol/protocol_http"
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
	apiPrefix := fmt.Sprintf("/api/%v/", p.GWConfig.Provider.Config.Get(KeyUsername))
	ph, err := httpProtocol.New(p.GWConfig, apiPrefix)
	if err != nil {
		return err
	}
	p.Gateway = ph
	return nil
}

// Close func
func (p *Provider) Close() error {
	return p.Gateway.Close()
}

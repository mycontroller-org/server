package tasmota

import (
	"fmt"

	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gateway_protocol"
	"github.com/mycontroller-org/backend/v2/plugin/gateway_protocol/mqtt"
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
	var err error
	switch p.GWConfig.Provider.ProtocolType {
	case gwpl.TypeMQTT:
		ms, _err := mqtt.New(p.GWConfig, rxMessageFunc)
		err = _err
		p.Gateway = ms
	default:
		return fmt.Errorf("Protocol not implemented: %s", p.GWConfig.Provider.ProtocolType)
	}
	return err
}

// Close func
func (p *Provider) Close() error {
	return p.Gateway.Close()
}

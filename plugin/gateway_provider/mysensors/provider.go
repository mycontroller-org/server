package mysensors

import (
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gateway_protocol"
	"github.com/mycontroller-org/backend/v2/plugin/gateway_protocol/mqtt"
	"github.com/mycontroller-org/backend/v2/plugin/gateway_protocol/serial"
)

// Provider implementation
type Provider struct {
	GWConfig *gwml.Config
	Gateway  gwpl.Gateway
}

// Post func
func (p *Provider) Post(rawMessage *msgml.RawMessage) error { return nil }

// Start func
func (p *Provider) Start(rxMessageFunc func(rawMsg *msgml.RawMessage) error) error {
	var err error
	switch p.GWConfig.Provider.ProtocolType {
	case gwpl.TypeMQTT:
		ms, _err := mqtt.New(p.GWConfig, rxMessageFunc)
		err = _err
		p.Gateway = ms
	case gwpl.TypeSerial:
		// update serial message splitter
		p.GWConfig.Provider.Config[serial.KeyMessageSplitter] = SerialMessageSplitter
		ms, _err := serial.New(p.GWConfig, rxMessageFunc)
		err = _err
		p.Gateway = ms
	}
	return err
}

// Close func
func (p *Provider) Close() error { return nil }

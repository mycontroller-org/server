package tasmota

import (
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	gwML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	utils "github.com/mycontroller-org/backend/v2/pkg/utils"
	gwPRL "github.com/mycontroller-org/backend/v2/plugin/gateway/protocol"
	mqtt "github.com/mycontroller-org/backend/v2/plugin/gateway/protocol/protocol_mqtt"
)

// Config of tasmota provider
type Config struct {
	Type     string
	Protocol cmap.CustomMap `json:"protocol"`
	// add provider configurations, if any
}

// Provider implementation
type Provider struct {
	Config        *Config
	GatewayConfig *gwML.Config
	Protocol      gwPRL.Protocol
	ProtocolType  string
}

// Init provider
func Init(gatewayConfig *gwML.Config) (*Provider, error) {
	cfg := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, gatewayConfig.Provider, cfg)
	if err != nil {
		return nil, err
	}
	provider := &Provider{
		Config:        cfg,
		GatewayConfig: gatewayConfig,
		ProtocolType:  cfg.Protocol.GetString(model.NameType),
	}
	return provider, nil
}

// Start func
func (p *Provider) Start(receivedMessageHandler func(rawMsg *msgML.RawMessage) error) error {
	var err error
	switch p.ProtocolType {
	case gwPRL.TypeMQTT:
		// update subscription topics
		protocol, _err := mqtt.New(p.GatewayConfig, p.Config.Protocol, receivedMessageHandler)
		err = _err
		p.Protocol = protocol
	default:
		return fmt.Errorf("protocol not implemented: %s", p.ProtocolType)
	}
	return err
}

// Close func
func (p *Provider) Close() error {
	return p.Protocol.Close()
}

// Post func
func (p *Provider) Post(msg *msgML.Message) error {
	rawMsg, err := p.ToRawMessage(msg)
	if err != nil {
		return err
	}
	return p.Protocol.Write(rawMsg)
}

package tasmota

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	utils "github.com/mycontroller-org/server/v2/pkg/utils"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	mqtt "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/protocol_mqtt"
	providerTY "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/type"
)

const PluginTasmota = "tasmota"

// Config of tasmota provider
type Config struct {
	Type     string
	Protocol cmap.CustomMap `json:"protocol"`
	// add provider configurations, if any
}

// Provider implementation
type Provider struct {
	Config        *Config
	GatewayConfig *gwTY.Config
	Protocol      gwPtl.Protocol
	ProtocolType  string
}

// NewPluginTasmota provider
func NewPluginTasmota(gatewayConfig *gwTY.Config) (providerTY.Plugin, error) {
	cfg := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, gatewayConfig.Provider, cfg)
	if err != nil {
		return nil, err
	}
	provider := &Provider{
		Config:        cfg,
		GatewayConfig: gatewayConfig,
		ProtocolType:  cfg.Protocol.GetString(types.NameType),
	}
	return provider, nil
}

func (p *Provider) Name() string {
	return PluginTasmota
}

// Start func
func (p *Provider) Start(receivedMessageHandler func(rawMsg *msgTY.RawMessage) error) error {
	var err error
	switch p.ProtocolType {
	case gwPtl.TypeMQTT:
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
func (p *Provider) Post(msg *msgTY.Message) error {
	rawMsg, err := p.ToRawMessage(msg)
	if err != nil {
		return err
	}
	return p.Protocol.Write(rawMsg)
}

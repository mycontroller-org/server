package generic

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	httpGenericProtocol "github.com/mycontroller-org/server/v2/plugin/gateway/provider/generic/protocol_http_generic"
	mqttGenericProtocol "github.com/mycontroller-org/server/v2/plugin/gateway/provider/generic/protocol_mqtt_generic"
	providerTY "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
)

const (
	PluginGeneric = "generic"
)

// Provider implementation
type Provider struct {
	Config        *Config
	GatewayConfig *gwTY.Config
	Protocol      GenericProtocol
	ProtocolType  string
}

// NewPluginGeneric provider
func NewPluginGeneric(gatewayConfig *gwTY.Config) (providerTY.Plugin, error) {
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
	return PluginGeneric
}

// Start func
func (p *Provider) Start(receivedMessageHandler func(rawMsg *msgTY.RawMessage) error) error {
	var err error
	switch p.ProtocolType {
	case gwPtl.TypeMQTT:
		// update subscription topics
		protocol, _err := mqttGenericProtocol.New(p.GatewayConfig, p.Config.Protocol, receivedMessageHandler)
		err = _err
		p.Protocol = protocol

	case gwPtl.TypeHttp:
		protocol, _err := httpGenericProtocol.New(p.GatewayConfig, p.Config.Protocol, receivedMessageHandler)
		err = _err
		p.Protocol = protocol

	default:
		return fmt.Errorf("protocol not implemented: %s", p.ProtocolType)
	}
	return err
}

// Close func
func (p *Provider) Close() error {
	if p.Protocol != nil {
		return p.Protocol.Close()
	}
	return nil
}

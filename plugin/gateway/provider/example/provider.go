package sample

import (
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	utils "github.com/mycontroller-org/backend/v2/pkg/utils"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gateway/protocol"
)

// Config of provider
type Config struct {
	Protocol cmap.CustomMap `json:"protocol"`
	// add provider configurations, if any
}

// Provider implementation
type Provider struct {
	Config        *Config
	GatewayConfig *gwml.Config
	Protocol      gwpl.Protocol
}

// Init MySensors provider
func Init(gatewayConfig *gwml.Config) *Provider {
	cfg := &Config{}
	utils.MapToStruct(utils.TagNameNone, gatewayConfig.Provider, cfg)
	provider := &Provider{
		Config:        cfg,
		GatewayConfig: gatewayConfig,
	}
	return provider
}

// Start func
func (p *Provider) Start(rxMessageFunc func(rawMsg *msgml.RawMessage) error) error {
	return nil
}

// Close func
func (p *Provider) Close() error {
	// do internal works
	// close gateway
	return p.Protocol.Close()
}

// Post func
func (p *Provider) Post(rawMsg *msgml.RawMessage) error {
	return p.Protocol.Write(rawMsg)
}

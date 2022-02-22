package generic

import (
	"fmt"

	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	"github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	mqtt "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/protocol_mqtt"
	providerTY "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

const (
	PluginGeneric           = "generic"
	ProtocolTypeHttpGeneric = "http_generic"

	schedulePrefix      = "generic_provider"
	defaultPoolInterval = "10m"
)

// Provider implementation
type Provider struct {
	Config            *Config
	GatewayConfig     *gwTY.Config
	Protocol          gwPtl.Protocol
	ProtocolType      string
	HttpProtocol      *HttpProtocol
	rawMessageHandler func(rawMsg *msgTY.RawMessage) error
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
	p.rawMessageHandler = receivedMessageHandler
	var err error
	switch p.ProtocolType {
	case gwPtl.TypeMQTT:
		// update subscription topics
		protocol, _err := mqtt.New(p.GatewayConfig, p.Config.Protocol, receivedMessageHandler)
		err = _err
		p.Protocol = protocol

	case ProtocolTypeHttpGeneric:
		if len(p.HttpProtocol.Addresses) == 0 {
			return nil
		}
		for index := range p.HttpProtocol.Addresses {
			cfg := p.HttpProtocol.Addresses[index]
			err = p.schedule(index, &cfg)
			if err != nil {
				zap.L().Error("error on schedule", zap.String("gatewayId", p.GatewayConfig.ID), zap.String("address", cfg.Address), zap.Error(err))
			}
		}

	default:
		return fmt.Errorf("protocol not implemented: %s", p.ProtocolType)
	}
	return err
}

// Close func
func (p *Provider) Close() error {
	p.unscheduleAll()
	if p.Protocol != nil {
		return p.Protocol.Close()
	}
	return nil
}

func (p *Provider) unscheduleAll() {
	coreScheduler.SVC.RemoveWithPrefix(fmt.Sprintf("%s_%s", schedulePrefix, p.GatewayConfig.ID))
}

func (p *Provider) schedule(index int, cfg *HttpConfig) error {
	if cfg.PoolInterval == "" {
		cfg.PoolInterval = defaultPoolInterval
	}

	triggerFunc := func() {
		rawMsg, err := p.executeHttpRequest(cfg)
		if err != nil {
			zap.L().Error("error on executing a request", zap.String("gatewayId", p.GatewayConfig.ID), zap.String("address", cfg.Address), zap.Error(err))
			return
		}
		if rawMsg != nil {
			err = p.rawMessageHandler(rawMsg)
			if err != nil {
				zap.L().Error("error on posting a rawmessage", zap.String("gatewayId", p.GatewayConfig.ID), zap.String("address", cfg.Address), zap.Error(err))
			}
		}
	}

	scheduleID := fmt.Sprintf("%s_%s_%d", schedulePrefix, p.GatewayConfig.ID, index)
	cronSpec := fmt.Sprintf("@every %s", cfg.PoolInterval)
	err := coreScheduler.SVC.AddFunc(scheduleID, cronSpec, triggerFunc)
	if err != nil {
		zap.L().Error("error on adding schedule", zap.Error(err))
		return err
	}
	zap.L().Debug("added a schedule", zap.String("schedulerID", scheduleID), zap.String("interval", cfg.PoolInterval))
	return nil
}

package tasmota

import (
	"context"
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	utils "github.com/mycontroller-org/server/v2/pkg/utils"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	mqtt "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/protocol_mqtt"
	providerTY "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

const (
	PluginTasmota = "tasmota"
	loggerName    = "gateway_tasmota"
)

// Config of tasmota provider
type Config struct {
	Type     string         `json:"type" yaml:"type"`
	Protocol cmap.CustomMap `json:"protocol" yaml:"protocol"`
	// add provider configurations, if any
}

// Provider implementation
type Provider struct {
	ctx              context.Context
	Config           *Config
	GatewayConfig    *gwTY.Config
	Protocol         gwPtl.Protocol
	ProtocolType     string
	logger           *zap.Logger
	bus              busTY.Plugin
	logRootDirectory string
}

// tasmota provider
func New(ctx context.Context, config *gwTY.Config) (providerTY.Plugin, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = utils.MapToStruct(utils.TagNameNone, config.Provider, cfg)
	if err != nil {
		return nil, err
	}
	provider := &Provider{
		ctx:              ctx,
		Config:           cfg,
		GatewayConfig:    config,
		ProtocolType:     cfg.Protocol.GetString(types.NameType),
		logger:           logger.Named(loggerName),
		bus:              bus,
		logRootDirectory: types.GetEnvString(types.ENV_DIR_GATEWAY_LOGS),
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
		protocol, _err := mqtt.New(p.logger, p.GatewayConfig, p.Config.Protocol, receivedMessageHandler, p.bus, p.logRootDirectory)
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

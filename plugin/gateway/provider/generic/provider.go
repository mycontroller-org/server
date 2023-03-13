package generic

import (
	"context"
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/types"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	httpGenericProtocol "github.com/mycontroller-org/server/v2/plugin/gateway/provider/generic/protocol_http_generic"
	mqttGenericProtocol "github.com/mycontroller-org/server/v2/plugin/gateway/provider/generic/protocol_mqtt_generic"
	providerTY "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

const (
	PluginGeneric = "generic"
	loggerName    = "gateway_generic"
)

// Provider implementation
type Provider struct {
	ctx              context.Context
	Config           *Config
	GatewayConfig    *gwTY.Config
	Protocol         GenericProtocol
	ProtocolType     string
	logger           *zap.Logger
	scheduler        schedulerTY.CoreScheduler
	bus              busTY.Plugin
	logRootDirectory string
}

// NewPluginGeneric provider
func NewPluginGeneric(ctx context.Context, config *gwTY.Config) (providerTY.Plugin, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	scheduler, err := schedulerTY.FromContext(ctx)
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
		scheduler:        scheduler,
		bus:              bus,
		logRootDirectory: types.GetEnvString(types.ENV_DIR_GATEWAY_LOGS),
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
		protocol, _err := mqttGenericProtocol.New(p.logger, p.GatewayConfig, p.Config.Protocol, receivedMessageHandler, p.bus, p.logRootDirectory)
		err = _err
		p.Protocol = protocol

	case gwPtl.TypeHttp:
		protocol, _err := httpGenericProtocol.New(p.logger, p.GatewayConfig, p.Config.Protocol, receivedMessageHandler, p.scheduler)
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

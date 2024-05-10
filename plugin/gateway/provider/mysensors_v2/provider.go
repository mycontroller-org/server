package mysensors

import (
	"context"
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	utils "github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	ethernet "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/protocol_ethernet"
	mqtt "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/protocol_mqtt"
	serial "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/protocol_serial"
	providerTY "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

const (
	PluginMySensorsV2 = "mysensors_v2"
	loggerName        = "gateway_mysensors_v2"
)

// Config of this provider
type Config struct {
	Type                     string         `json:"type" yaml:"type"`
	EnableStreamMessageAck   bool           `json:"enableStreamMessageAck" yaml:"enableStreamMessageAck"`
	EnableInternalMessageAck bool           `json:"enableInternalMessageAck" yaml:"enableInternalMessageAck"`
	RetryCount               int            `json:"retryCount" yaml:"retryCount"`
	Timeout                  string         `json:"timeout" yaml:"timeout"`
	Protocol                 cmap.CustomMap `json:"protocol" yaml:"protocol"`
}

// Provider implementation
type Provider struct {
	Config           *Config
	GatewayConfig    *gwTY.Config
	Protocol         gwPtl.Protocol
	ProtocolType     string
	logger           *zap.Logger
	scheduler        schedulerTY.CoreScheduler
	bus              busTY.Plugin
	logRootDirectory string
}

const (
	defaultTimeout             = time.Millisecond * 200
	timeoutAllowedMinimumLevel = time.Millisecond * 10
)

// MySensors v2 provider
func New(ctx context.Context, config *gwTY.Config) (providerTY.Plugin, error) {
	logger, err := loggerUtils.FromContext(ctx)
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
		Config:           cfg,
		GatewayConfig:    config,
		ProtocolType:     cfg.Protocol.GetString(types.NameType),
		logger:           logger.Named(loggerName),
		scheduler:        scheduler,
		bus:              bus,
		logRootDirectory: types.GetEnvString(types.ENV_DIR_GATEWAY_LOGS),
	}
	provider.logger.Debug("Config details", zap.Any("received", config.Provider), zap.Any("converted", cfg))
	return provider, nil
}

func (p *Provider) Name() string {
	return PluginMySensorsV2
}

// Start func
func (p *Provider) Start(receivedMessageHandler func(rawMsg *msgTY.RawMessage) error) error {
	var err error
	switch p.ProtocolType {
	case gwPtl.TypeMQTT:
		protocol, _err := mqtt.New(p.logger, p.GatewayConfig, p.Config.Protocol, receivedMessageHandler, p.bus, p.logRootDirectory)
		err = _err
		p.Protocol = protocol

	case gwPtl.TypeSerial:
		// update serial message splitter
		p.Config.Protocol.Set(serial.KeyMessageSplitter, serialMessageSplitter, nil)
		protocol, _err := serial.New(p.logger, p.GatewayConfig, p.Config.Protocol, receivedMessageHandler, p.bus, p.logRootDirectory)
		err = _err
		p.Protocol = protocol

	case gwPtl.TypeEthernet:
		// update ethernet message splitter
		p.Config.Protocol.Set(serial.KeyMessageSplitter, ethernetMessageSplitter, nil)
		protocol, _err := ethernet.New(p.logger, p.GatewayConfig, p.Config.Protocol, receivedMessageHandler, p.bus, p.logRootDirectory)
		err = _err
		p.Protocol = protocol

	}

	if err != nil {
		return err
	}

	// load firmware purge job
	firmwarePurgeJobName := fmt.Sprintf("%s_%s", firmwarePurgeJobName, p.GatewayConfig.ID)
	err = p.scheduler.AddFunc(firmwarePurgeJobName, firmwarePurgeJobCron, firmwareRawPurge)
	if err != nil {
		return err
	}
	err = p.initEventListener(p.GatewayConfig.ID)
	if err != nil {
		return err
	}

	return p.scheduleNodeDiscover()
}

// Close func
func (p *Provider) Close() error {
	// remove all schedules on this gateway
	p.scheduler.RemoveWithPrefix(fmt.Sprintf("%s_%s", schedulePrefix, p.GatewayConfig.ID))

	// stop event listener
	p.closeEventListener()

	// remove firmware purge job
	fwPurgeJobName := fmt.Sprintf("%s_%s", firmwarePurgeJobName, p.GatewayConfig.ID)
	p.scheduler.RemoveFunc(fwPurgeJobName)
	// close gateway
	return p.Protocol.Close()
}

// Post func
// returns the status and error message if any
func (p *Provider) Post(msg *msgTY.Message) error {
	rawMsg, err := p.toRawMessage(msg)
	if err != nil {
		p.logger.Error("error on converting to raw message", zap.String("gatewayId", p.GatewayConfig.ID), zap.Error(err))
		return err
	}

	if rawMsg == nil {
		return nil
	}

	// if acknowledge not enabled
	if !rawMsg.IsAckEnabled {
		return p.Protocol.Write(rawMsg)
	}

	// if acknowledge enabled
	// wait for acknowledgement message
	// keeping channel capacity leads infinite wait on a situation like, receiving more than one ack
	// to address this, as a workaround changing the channel capacity from 0 to some defined numbers
	// TODO: still it is possible to get in to infinite wait lock, when it receives defined number of ack
	ackChannel := concurrency.NewChannel(20)
	ackTopic := fmt.Sprintf("%s.%s", p.GatewayConfig.ID, rawMsg.ID)
	ackSubscriptionID, err := p.bus.Subscribe(
		ackTopic,
		func(event *busTY.BusData) {
			p.logger.Debug("acknowledgement status", zap.Any("event", event))
			ackChannel.SafeSend(true)
		},
	)
	if err != nil {
		return err
	}

	// on exit unsubscribe and close the channel
	defer func() {
		err := p.bus.Unsubscribe(ackTopic, ackSubscriptionID)
		if err != nil {
			p.logger.Error("error on unsubscribe", zap.Error(err), zap.String("topic", ackTopic), zap.Any("sunscriptionID", ackSubscriptionID))
		}
		ackChannel.SafeClose()
	}()

	timeout, err := time.ParseDuration(p.Config.Timeout)
	if err != nil {
		// failed to parse timeout, running with default
		timeout = defaultTimeout
		p.logger.Warn("failed to parse timeout, running with default timeout", zap.String("timeout", p.Config.Timeout), zap.String("default", defaultTimeout.String()), zap.Error(err))
	}

	// minimum timeout
	if timeout < timeoutAllowedMinimumLevel {
		p.logger.Info("adjusting timeout to mimimum allowed level", zap.String("supplied", timeout.String()), zap.String("minimum", timeoutAllowedMinimumLevel.String()))
		timeout = timeoutAllowedMinimumLevel
	}

	retryCount := p.Config.RetryCount
	// check retry count, and reset if invalid number set
	if retryCount < 1 {
		retryCount = 1
		p.logger.Info("adjusting retry count", zap.Int("supplied", p.Config.RetryCount), zap.Int("updated", retryCount))
	}

	messageSent := false
	startTime := time.Now()
	for retry := 1; retry <= retryCount; retry++ {
		// write into gateway
		err = p.Protocol.Write(rawMsg)
		if err != nil {
			return err
		}

		// wait till timeout or acknowledgement, which one comes earlier
		select {
		case <-ackChannel.CH:
			messageSent = true
		case <-time.After(timeout):
			// wait till timeout
		}
		if messageSent {
			break
		}
	}
	if messageSent {
		return nil
	}
	return fmt.Errorf("no acknowledgement received, tried maximum retries. retryCount:%d, timeTaken:%s", retryCount, time.Since(startTime).String())
}

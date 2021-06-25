package mysensors

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/model"
	busML "github.com/mycontroller-org/server/v2/pkg/model/bus"
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	gwML "github.com/mycontroller-org/server/v2/pkg/model/gateway"
	msgML "github.com/mycontroller-org/server/v2/pkg/model/message"
	sch "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	utils "github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	gwpl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	ethernet "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/protocol_ethernet"
	mqtt "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/protocol_mqtt"
	serial "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/protocol_serial"
	"go.uber.org/zap"
)

// Config of this provider
type Config struct {
	Type                     string         `json:"type"`
	EnableStreamMessageAck   bool           `json:"enableStreamMessageAck"`
	EnableInternalMessageAck bool           `json:"enableInternalMessageAck"`
	RetryCount               int            `json:"retryCount"`
	Timeout                  string         `json:"timeout"`
	Protocol                 cmap.CustomMap `json:"protocol"`
}

// Provider implementation
type Provider struct {
	Config        *Config
	GatewayConfig *gwML.Config
	Protocol      gwpl.Protocol
	ProtocolType  string
}

const (
	defaultTimeout             = time.Millisecond * 200
	timeoutAllowedMinimumLevel = time.Millisecond * 10
)

// Init MySensors provider
func Init(gatewayCfg *gwML.Config) (*Provider, error) {
	cfg := &Config{}
	err := utils.MapToStruct(utils.TagNameNone, gatewayCfg.Provider, cfg)
	if err != nil {
		return nil, err
	}
	provider := &Provider{
		Config:        cfg,
		GatewayConfig: gatewayCfg,
		ProtocolType:  cfg.Protocol.GetString(model.NameType),
	}
	zap.L().Debug("Config details", zap.Any("received", gatewayCfg.Provider), zap.Any("converted", cfg))
	return provider, nil
}

// Start func
func (p *Provider) Start(receivedMessageHandler func(rawMsg *msgML.RawMessage) error) error {
	var err error
	switch p.ProtocolType {
	case gwpl.TypeMQTT:
		protocol, _err := mqtt.New(p.GatewayConfig, p.Config.Protocol, receivedMessageHandler)
		err = _err
		p.Protocol = protocol

	case gwpl.TypeSerial:
		// update serial message splitter
		p.Config.Protocol.Set(serial.KeyMessageSplitter, serialMessageSplitter, nil)
		protocol, _err := serial.New(p.GatewayConfig, p.Config.Protocol, receivedMessageHandler)
		err = _err
		p.Protocol = protocol

	case gwpl.TypeEthernet:
		// update ethernet message splitter
		p.Config.Protocol.Set(serial.KeyMessageSplitter, ethernetMessageSplitter, nil)
		protocol, _err := ethernet.New(p.GatewayConfig, p.Config.Protocol, receivedMessageHandler)
		err = _err
		p.Protocol = protocol

	}

	if err != nil {
		return err
	}

	// load firmware purge job
	firmwarePurgeJobName := fmt.Sprintf("%s_%s", firmwarePurgeJobName, p.GatewayConfig.ID)
	err = sch.SVC.AddFunc(firmwarePurgeJobName, firmwarePurgeJobCron, firmwareRawPurge)
	if err != nil {
		return err
	}
	return initEventListener(p.GatewayConfig.ID)
}

// Close func
func (p *Provider) Close() error {
	// stop event listener
	closeEventListener()

	// remove firmware purge job
	fwPurgeJobName := fmt.Sprintf("%s_%s", firmwarePurgeJobName, p.GatewayConfig.ID)
	sch.SVC.RemoveFunc(fwPurgeJobName)
	// close gateway
	return p.Protocol.Close()
}

// Post func
// returns the status and error message if any
func (p *Provider) Post(msg *msgML.Message) error {
	rawMsg, err := p.toRawMessage(msg)
	if err != nil {
		zap.L().Error("error on converting to raw message", zap.Error(err), zap.String("gatewayId", p.GatewayConfig.ID))
		return err
	}

	if rawMsg == nil {
		return nil
	}

	// if acknowledge not enabled
	if !rawMsg.IsAckEnabled {
		err := p.Protocol.Write(rawMsg)
		if err != nil {
			return err
		}
		return nil
	}

	// if acknowledge enabled
	// wait for acknowledgement message
	ackChannel := concurrency.NewChannel(0)
	ackTopic := mcbus.GetTopicPostRawMessageAcknowledgement(p.GatewayConfig.ID, rawMsg.ID)
	ackSubscriptionID, err := mcbus.Subscribe(
		ackTopic,
		func(event *busML.BusData) {
			zap.L().Debug("acknowledgement status", zap.Any("event", event))
			ackChannel.SafeSend(true)
		},
	)
	if err != nil {
		return err
	}

	// on exit unsubscribe and close the channel
	defer func() {
		err := mcbus.Unsubscribe(ackTopic, ackSubscriptionID)
		if err != nil {
			zap.L().Error("error on unsubscribe", zap.Error(err), zap.String("topic", ackTopic), zap.Any("sunscriptionID", ackSubscriptionID))
		}
		ackChannel.SafeClose()
	}()

	timeout, err := time.ParseDuration(p.Config.Timeout)
	if err != nil {
		// failed to parse timeout, running with default
		timeout = defaultTimeout
		zap.L().Warn("failed to parse timeout, running with default timeout", zap.String("timeout", p.Config.Timeout), zap.String("default", defaultTimeout.String()), zap.Error(err))
	}

	// minimum timeout
	if timeout < timeoutAllowedMinimumLevel {
		zap.L().Info("adjusting timeout to mimimum allowed level", zap.String("supplied", timeout.String()), zap.String("minimum", timeoutAllowedMinimumLevel.String()))
		timeout = timeoutAllowedMinimumLevel
	}

	retryCount := p.Config.RetryCount
	// check retry count, and reset if invalid number set
	if retryCount < 1 {
		retryCount = 1
		zap.L().Info("adjusting retry count", zap.Int("supplied", p.Config.RetryCount), zap.Int("updated", retryCount))
	}

	messageSent := false
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
	return fmt.Errorf("no acknowledgement received, tried maximum retries. retryCount:%d", retryCount)
}

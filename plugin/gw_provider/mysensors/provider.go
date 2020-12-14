package mysensors

import (
	"fmt"
	"time"

	"github.com/mustafaturan/bus"
	"github.com/mycontroller-org/backend/v2/pkg/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	utils "github.com/mycontroller-org/backend/v2/pkg/utils"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gw_protocol"
	mqtt "github.com/mycontroller-org/backend/v2/plugin/gw_protocol/protocol_mqtt"
	serial "github.com/mycontroller-org/backend/v2/plugin/gw_protocol/protocol_serial"
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
	GatewayConfig *gwml.Config
	Protocol      gwpl.Protocol
	ProtocolType  string
}

const (
	defaultTimeout             = time.Millisecond * 200
	timeoutAllowedMinimumLevel = time.Millisecond * 10
)

// Init MySensors provider
func Init(gatewayCfg *gwml.Config) *Provider {
	cfg := &Config{}
	utils.MapToStruct(utils.TagNameNone, gatewayCfg.Provider, cfg)
	provider := &Provider{
		Config:        cfg,
		GatewayConfig: gatewayCfg,
		ProtocolType:  cfg.Protocol.GetString(model.NameType),
	}
	zap.L().Debug("Config details", zap.Any("received", gatewayCfg.Provider), zap.Any("converted", cfg))
	return provider
}

// Start func
func (p *Provider) Start(receivedMessageHandler func(rawMsg *msgml.RawMessage) error) error {
	var err error
	switch p.ProtocolType {
	case gwpl.TypeMQTT:
		protocol, _err := mqtt.New(p.GatewayConfig, p.Config.Protocol, receivedMessageHandler)
		err = _err
		p.Protocol = protocol
	case gwpl.TypeSerial:
		// update serial message splitter
		p.GatewayConfig.Provider.Set(serial.KeyMessageSplitter, serialMessageSplitter, nil)
		protocol, _err := serial.New(p.GatewayConfig, receivedMessageHandler)
		err = _err
		p.Protocol = protocol
	}
	if err != nil {
		return err
	}

	// load firmware purge job
	firmwarePurgeJobName := fmt.Sprintf("%s_%s", firmwarePurgeJobName, p.GatewayConfig.ID)
	return svc.SCH.AddFunc(firmwarePurgeJobName, firmwarePurgeJobCron, fwStore.purge)
}

// Close func
func (p *Provider) Close() error {
	// remove firmware purge job
	fwPurgeJobName := fmt.Sprintf("%s_%s", firmwarePurgeJobName, p.GatewayConfig.ID)
	svc.SCH.RemoveFunc(fwPurgeJobName)
	// close gateway
	return p.Protocol.Close()
}

// Post func
// returns the status and error message if any
func (p *Provider) Post(rawMsg *msgml.RawMessage) error {

	// if acknowledge not enabled
	if !rawMsg.AcknowledgeEnabled {
		err := p.Protocol.Write(rawMsg)
		if err != nil {
			return err
		}
		return nil
	}

	// if acknowledge enabled
	// wait for acknowledgement message
	ackChannel := make(chan bool, 1)
	ackTopic := getAcknowledgementTopic(p.GatewayConfig.ID, rawMsg.ID)
	mcbus.Subscribe(ackTopic, &bus.Handler{
		Matcher: ackTopic,
		Handle: func(e *bus.Event) {
			zap.L().Debug("acknowledgement status", zap.Any("event", e))
			// TODO: facing issue, closed channel. Fix this
			ackChannel <- true
		},
	})

	// on exit unsubscribe and close the channel
	defer func() {
		mcbus.Unsubscribe(ackTopic)
		close(ackChannel)
	}()

	timeout, err := time.ParseDuration(p.Config.Timeout)
	if err != nil {
		// failed to parse timeout, running with default
		timeout = defaultTimeout
		zap.L().Warn("Failed to parse timeout, running with default timeout", zap.String("timeout", p.Config.Timeout), zap.String("default", defaultTimeout.String()), zap.Error(err))
	}

	// minimum timeout
	if timeout < timeoutAllowedMinimumLevel {
		zap.L().Info("adjesting timeout to mimimum allowed level", zap.String("supplied", timeout.String()), zap.String("minimum", timeoutAllowedMinimumLevel.String()))
		timeout = timeoutAllowedMinimumLevel
	}

	retryCount := p.Config.RetryCount
	// check retry count, and reset if invalid number set
	if retryCount < 1 {
		retryCount = 1
		zap.L().Info("adjesting retry count", zap.Int("supplied", p.Config.RetryCount), zap.Int("updated", retryCount))
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
		case <-ackChannel:
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
	return fmt.Errorf("No acknowledgement received, tried maximum retries. retryCount:%d", retryCount)
}

package mqtt

import (
	"crypto/tls"
	"fmt"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	gwptcl "github.com/mycontroller-org/backend/v2/plugin/gateway/protocol"
	msglogger "github.com/mycontroller-org/backend/v2/plugin/gateway/protocol/message_logger"

	"go.uber.org/zap"
)

// Constants in serial gateway
const (
	transmitPreDelayDefault = time.Microsecond * 1 // 1 micro second
	reconnectDelayDefault   = time.Second * 10     // 10 seconds
)

// Config data
type Config struct {
	Type               string
	Broker             string
	Username           string
	Password           string
	Subscribe          string
	Publish            string
	QoS                int
	TransmitPreDelay   string
	InsecureSkipVerify bool
}

// Endpoint data
type Endpoint struct {
	GatewayCfg     *gwml.Config
	Config         Config
	receiveMsgFunc func(rm *msgml.RawMessage) error
	Client         paho.Client
	messageLogger  msglogger.MessageLogger
	txPreDelay     time.Duration
}

// New mqtt driver
func New(gwCfg *gwml.Config, protocol cmap.CustomMap, rxMsgFunc func(rm *msgml.RawMessage) error) (*Endpoint, error) {
	zap.L().Debug("Init protocol", zap.String("gateway", gwCfg.ID))
	start := time.Now()
	cfg := Config{}
	err := utils.MapToStruct(utils.TagNameNone, protocol, &cfg)
	if err != nil {
		return nil, err
	}

	// endpoint
	endpoint := &Endpoint{
		GatewayCfg:     gwCfg,
		Config:         cfg,
		receiveMsgFunc: rxMsgFunc,
		txPreDelay:     utils.ToDuration(cfg.TransmitPreDelay, transmitPreDelayDefault),
	}

	// add void logger to avoid nill exception, till er get successful connection
	endpoint.messageLogger = msglogger.GetVoidLogger()

	opts := paho.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetClientID(utils.RandID())
	opts.SetCleanSession(false)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetryInterval(utils.ToDuration(gwCfg.ReconnectDelay, reconnectDelayDefault))
	opts.SetOnConnectHandler(endpoint.onConnectionHandler)
	opts.SetConnectionLostHandler(endpoint.onConnectionLostHandler)

	// update tls config
	tlsConfig := &tls.Config{InsecureSkipVerify: cfg.InsecureSkipVerify}
	opts.SetTLSConfig(tlsConfig)

	c := paho.NewClient(opts)
	token := c.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		return nil, err
	}

	// init and start actual message message logger
	endpoint.messageLogger = msglogger.Init(gwCfg.ID, gwCfg.MessageLogger, messageFormatter)
	endpoint.messageLogger.Start()

	// adding client
	endpoint.Client = c

	err = endpoint.Subscribe(cfg.Subscribe)
	if err != nil {
		zap.L().Error("Failed to subscribe a topic", zap.String("topic", cfg.Subscribe), zap.Error(err))
	}
	zap.L().Debug("MQTT client connected successfully", zap.String("timeTaken", time.Since(start).String()), zap.Any("clientConfig", cfg))
	return endpoint, nil
}

// messageFormatter returns the message as string format
func messageFormatter(rawMsg *msgml.RawMessage) string {
	direction := "Sent"
	if rawMsg.IsReceived {
		direction = "Recd"
	}
	return fmt.Sprintf(
		"%v\t%s\t%v\t\t\t%s\n",
		rawMsg.Timestamp.Format("2006-01-02T15:04:05.000Z0700"),
		direction,
		rawMsg.Others.Get(gwptcl.KeyMqttTopic),
		string(rawMsg.Data),
	)
}

func (ep *Endpoint) onConnectionHandler(c paho.Client) {
	zap.L().Debug("MQTT connection success", zap.Any("gateway", ep.GatewayCfg.ID))
	state := model.State{
		Status:  model.StatusUp,
		Message: "Connected successfully",
		Since:   time.Now(),
	}
	busUtils.SetGatewayState(ep.GatewayCfg.ID, state)
}

func (ep *Endpoint) onConnectionLostHandler(c paho.Client, err error) {
	zap.L().Error("MQTT connection lost", zap.Any("gateway", ep.GatewayCfg.ID), zap.Error(err))
	state := model.State{
		Status:  model.StatusDown,
		Message: err.Error(),
		Since:   time.Now(),
	}
	busUtils.SetGatewayState(ep.GatewayCfg.ID, state)
}

// Write publishes a payload
func (ep *Endpoint) Write(rawMsg *msgml.RawMessage) error {
	zap.L().Debug("About to send a message", zap.Any("rawMessage", rawMsg))
	topics := rawMsg.Others.Get(gwptcl.KeyMqttTopic).([]string)
	qos := byte(ep.Config.QoS)
	rawMsg.IsReceived = false

	time.Sleep(ep.txPreDelay) // transmit pre delay

	for _, t := range topics {
		_topic := fmt.Sprintf("%s/%s", ep.Config.Publish, t)
		rawMsgCloned := rawMsg.Clone()
		rawMsgCloned.Others.Set(gwptcl.KeyMqttTopic, _topic, nil)
		rawMsgCloned.Timestamp = time.Now()
		ep.messageLogger.AsyncWrite(rawMsgCloned)

		token := ep.Client.Publish(_topic, qos, false, rawMsg.Data)
		if token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}

// Close the driver
func (ep *Endpoint) Close() error {
	if ep.Client.IsConnected() {
		ep.Client.Unsubscribe(ep.Config.Subscribe)
		ep.Client.Disconnect(0)
		zap.L().Debug("MQTT Client connection closed", zap.String("gateway", ep.GatewayCfg.ID))
	}
	ep.messageLogger.Close()
	return nil
}

func (ep *Endpoint) getCallBack() func(paho.Client, paho.Message) {
	return func(c paho.Client, message paho.Message) {
		rawMsg := msgml.NewRawMessage(true, message.Payload())
		rawMsg.Others.Set(gwptcl.KeyMqttTopic, message.Topic(), nil)
		rawMsg.Others.Set(gwptcl.KeyMqttQoS, int(message.Qos()), nil)

		ep.messageLogger.AsyncWrite(rawMsg)
		err := ep.receiveMsgFunc(rawMsg)
		if err != nil {
			zap.L().Error("Failed to process", zap.String("gateway", ep.GatewayCfg.ID), zap.Any("rawMessage", rawMsg), zap.Error(err))
		}
	}
}

// Subscribe a topic
func (ep *Endpoint) Subscribe(topic string) error {
	token := ep.Client.Subscribe(topic, 0, ep.getCallBack())
	token.WaitTimeout(3 * time.Second)
	return token.Error()
}

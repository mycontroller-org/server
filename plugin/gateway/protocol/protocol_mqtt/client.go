package mqtt

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	msglogger "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/message_logger"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"

	"go.uber.org/zap"
)

// Constants in serial gateway
const (
	transmitPreDelayDefault = time.Microsecond * 1 // 1 micro second
	reconnectDelayDefault   = time.Second * 10     // 10 seconds
)

// Config data
type Config struct {
	Type             string
	Broker           string
	Username         string
	Password         string
	Subscribe        string
	Publish          string
	QoS              int
	TransmitPreDelay string
	Insecure         bool
}

// Endpoint data
type Endpoint struct {
	GatewayCfg     *gwTY.Config
	Config         Config
	receiveMsgFunc func(rm *msgTY.RawMessage) error
	Client         paho.Client
	messageLogger  msglogger.MessageLogger
	txPreDelay     time.Duration
}

// New mqtt driver
func New(gwCfg *gwTY.Config, protocol cmap.CustomMap, rxMsgFunc func(rm *msgTY.RawMessage) error) (*Endpoint, error) {
	zap.L().Debug("making mqtt connection", zap.String("gatewayId", gwCfg.ID))
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
	opts.SetClientID(fmt.Sprintf("%s-%s", gwCfg.ID, utils.RandIDWithLength(5)))
	opts.SetCleanSession(false)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetryInterval(utils.ToDuration(gwCfg.ReconnectDelay, reconnectDelayDefault))
	opts.SetOnConnectHandler(endpoint.onConnectionHandler)
	opts.SetConnectionLostHandler(endpoint.onConnectionLostHandler)

	// update tls config
	tlsConfig := &tls.Config{InsecureSkipVerify: cfg.Insecure}
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

	zap.L().Debug("MQTT client connected successfully", zap.String("timeTaken", time.Since(start).String()), zap.String("gatewayId", gwCfg.ID), zap.Any("clientConfig", cfg))
	return endpoint, nil
}

// messageFormatter returns the message as string format
func messageFormatter(rawMsg *msgTY.RawMessage) string {
	direction := "Sent"
	if rawMsg.IsReceived {
		direction = "Recd"
	}
	return fmt.Sprintf(
		"%v\t%s\t%v\t\t\t%s\n",
		rawMsg.Timestamp.Format("2006-01-02T15:04:05.000Z0700"),
		direction,
		rawMsg.Others.Get(gwPtl.KeyMqttTopic),
		convertor.ToString(rawMsg.Data),
	)
}

func (ep *Endpoint) onConnectionHandler(c paho.Client) {
	zap.L().Debug("MQTT connection success", zap.Any("gatewayId", ep.GatewayCfg.ID))
	state := types.State{
		Status:  types.StatusUp,
		Message: "Connected successfully",
		Since:   time.Now(),
	}

	err := ep.Subscribe(ep.Config.Subscribe)
	if err != nil {
		zap.L().Error("failed to subscribe topics", zap.String("gatewayId", ep.GatewayCfg.ID), zap.String("topics", ep.Config.Subscribe), zap.Error(err))
		state.Message = fmt.Sprintf("Connected successfully, error on subscription:%s", err.Error())
	}

	busUtils.SetGatewayState(ep.GatewayCfg.ID, state)
}

func (ep *Endpoint) onConnectionLostHandler(c paho.Client, err error) {
	zap.L().Error("mqtt connection lost", zap.Any("gatewayId", ep.GatewayCfg.ID), zap.Error(err))
	state := types.State{
		Status:  types.StatusDown,
		Message: err.Error(),
		Since:   time.Now(),
	}
	busUtils.SetGatewayState(ep.GatewayCfg.ID, state)
}

// Write publishes a payload
func (ep *Endpoint) Write(rawMsg *msgTY.RawMessage) error {
	zap.L().Debug("About to send a message", zap.String("gatewayId", ep.GatewayCfg.ID), zap.Any("rawMessage", rawMsg))
	topics := rawMsg.Others.Get(gwPtl.KeyMqttTopic).([]string)
	qos := byte(ep.Config.QoS)
	rawMsg.IsReceived = false

	time.Sleep(ep.txPreDelay) // transmit pre delay

	for _, t := range topics {
		_topic := fmt.Sprintf("%s/%s", ep.Config.Publish, t)
		rawMsgCloned := rawMsg.Clone()
		rawMsgCloned.Others.Set(gwPtl.KeyMqttTopic, _topic, nil)
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
		zap.L().Debug("MQTT Client connection closed", zap.String("gatewayId", ep.GatewayCfg.ID))
	}
	ep.messageLogger.Close()
	return nil
}

func (ep *Endpoint) getCallBack() func(paho.Client, paho.Message) {
	return func(c paho.Client, message paho.Message) {
		rawMsg := msgTY.NewRawMessage(true, message.Payload())
		rawMsg.Others.Set(gwPtl.KeyMqttTopic, message.Topic(), nil)
		rawMsg.Others.Set(gwPtl.KeyMqttQoS, int(message.Qos()), nil)

		ep.messageLogger.AsyncWrite(rawMsg)
		err := ep.receiveMsgFunc(rawMsg)
		if err != nil {
			zap.L().Error("failed to process received message", zap.String("gatewayId", ep.GatewayCfg.ID), zap.Any("rawMessage", rawMsg), zap.Error(err))
		}
	}
}

// Subscribe topics
// supply comma separated topic names
// example: root/topic1/hello,root/topic2/#
func (ep *Endpoint) Subscribe(topicStr string) error {
	topics := strings.Split(topicStr, ",")
	for _, topic := range topics {
		topic = strings.TrimSpace(topic)
		token := ep.Client.Subscribe(topic, 0, ep.getCallBack())
		token.WaitTimeout(3 * time.Second)
		if token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}

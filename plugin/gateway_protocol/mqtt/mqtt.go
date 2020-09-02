package mqtt

import (
	"fmt"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
	gwptcl "github.com/mycontroller-org/backend/v2/plugin/gateway_protocol"

	"go.uber.org/zap"
)

// Config data
type Config struct {
	Broker           string `json:"broker"`
	Username         string `json:"username"`
	Password         string `json:"-"`
	Subscribe        string `json:"subscribe"`
	Publish          string `json:"publish"`
	QoS              int    `json:"qos"`
	TransmitPreDelay string `json:"transmitPreDelay"`
}

// Endpoint data
type Endpoint struct {
	GwCfg          *gwml.Config
	Config         Config
	receiveMsgFunc func(rm *msgml.RawMessage) error
	Client         paho.Client
	rawMsgLogger   *gwptcl.RawMessageLogger
	txPreDelay     *time.Duration
}

// New mqtt driver
func New(gwCfg *gwml.Config, rxMsgFunc func(rm *msgml.RawMessage) error) (*Endpoint, error) {
	start := time.Now()
	cfg := Config{}
	err := ut.MapToStruct(ut.TagNameNone, gwCfg.Provider.Config, &cfg)
	if err != nil {
		return nil, err
	}

	var txPreDelay *time.Duration
	// get transmit pre delay
	if cfg.TransmitPreDelay != "" {
		_txPreDelay, err := time.ParseDuration(cfg.TransmitPreDelay)
		if err != nil {
			zap.L().Error("Failed to parse transmit delay", zap.String("transmitPreDelay", cfg.TransmitPreDelay))
		}
		txPreDelay = &_txPreDelay
	}

	msgFormatterFn := func(rawMsg *msgml.RawMessage) string {
		direction := "Tx"
		if rawMsg.IsReceived {
			direction = "Rx"
		}
		return fmt.Sprintf("%v\t%s\t%v\t\t%s\n", rawMsg.Timestamp.Format("2006-01-02T15:04:05.000Z0700"), direction, rawMsg.Others.Get(gwptcl.KeyMqttTopic), string(rawMsg.Data))
	}

	// endpoint
	d := &Endpoint{
		GwCfg:          gwCfg,
		Config:         cfg,
		receiveMsgFunc: rxMsgFunc,
		rawMsgLogger:   &gwptcl.RawMessageLogger{Config: gwCfg, MsgFormatterFn: msgFormatterFn},
		txPreDelay:     txPreDelay,
	}

	opts := paho.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetClientID(ut.RandID())
	opts.SetCleanSession(false)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(d.onConnectionHandler)
	opts.SetConnectionLostHandler(d.onConnectionLostHandler)

	c := paho.NewClient(opts)
	token := c.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		return nil, err
	}

	// adding client
	d.Client = c

	// start raw message logger
	d.rawMsgLogger.Start()

	err = d.Subscribe(cfg.Subscribe)
	if err != nil {
		zap.L().Error("Failed to subscribe a topic", zap.String("topic", cfg.Subscribe), zap.Error(err))
	}
	zap.L().Debug("MQTT client connected successfully", zap.String("timeTaken", time.Since(start).String()), zap.Any("clientConfig", cfg))
	return d, nil
}

func (d *Endpoint) onConnectionHandler(c paho.Client) {
	zap.L().Debug("MQTT connection success", zap.Any("gateway", d.GwCfg.Name))
}

func (d *Endpoint) onConnectionLostHandler(c paho.Client, err error) {
	zap.L().Error("MQTT connection lost", zap.Any("gateway", d.GwCfg.Name), zap.Error(err))
}

// Write publishes a payload
func (d *Endpoint) Write(rawMsg *msgml.RawMessage) error {
	zap.L().Debug("About to send a message", zap.Any("rawMessage", rawMsg))
	topics := rawMsg.Others.Get(gwptcl.KeyMqttTopic).([]string)
	qos := byte(d.Config.QoS)
	rawMsg.IsReceived = false
	// add transmit pre delay
	if d.txPreDelay != nil {
		time.Sleep(*d.txPreDelay)
	}
	for _, t := range topics {
		_topic := fmt.Sprintf("%s/%s", d.Config.Publish, t)
		rawMsgCloned := rawMsg.Clone()
		rawMsgCloned.Others.Set(gwptcl.KeyMqttTopic, _topic, nil)
		rawMsgCloned.Timestamp = time.Now()
		d.rawMsgLogger.AsyncWrite(rawMsgCloned)

		token := d.Client.Publish(_topic, qos, false, rawMsg.Data)
		if token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}

// Close the driver
func (d *Endpoint) Close() error {
	if d.Client.IsConnected() {
		d.Client.Unsubscribe(d.Config.Subscribe)
		d.Client.Disconnect(0)
		zap.L().Debug("MQTT Client connection closed", zap.String("gateway", d.GwCfg.Name))
	}
	d.rawMsgLogger.Close()
	return nil
}

func (d *Endpoint) getCallBack() func(paho.Client, paho.Message) {
	return func(c paho.Client, message paho.Message) {
		rawMsg := &msgml.RawMessage{
			Data:      message.Payload(),
			Timestamp: time.Now(),
			Others: map[string]interface{}{
				gwptcl.KeyMqttTopic: message.Topic(),
				gwptcl.KeyMqttQoS:   int(message.Qos()),
			},
		}
		rawMsg.IsReceived = true
		d.rawMsgLogger.AsyncWrite(rawMsg.Clone())
		err := d.receiveMsgFunc(rawMsg)
		if err != nil {
			zap.L().Error("Failed to send a raw message to queue", zap.String("gateway", d.GwCfg.Name), zap.Any("rawMessage", rawMsg), zap.Error(err))
		}
	}
}

// Subscribe a topic
func (d *Endpoint) Subscribe(topic string) error {
	token := d.Client.Subscribe(topic, 0, d.getCallBack())
	token.WaitTimeout(3 * time.Second)
	return token.Error()
}

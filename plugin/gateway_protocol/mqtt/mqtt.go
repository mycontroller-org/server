package mqtt

import (
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
	Broker    string `json:"broker"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	Subscribe string `json:"subscribe"`
	Publish   string `json:"publish"`
	QoS       int    `json:"qos"`
}

// Endpoint data
type Endpoint struct {
	GwCfg          *gwml.Config
	Config         Config
	receiveMsgFunc func(rm *msgml.RawMessage) error
	Client         paho.Client
}

// New mqtt driver
func New(gwCfg *gwml.Config, rxMsgFunc func(rm *msgml.RawMessage) error) (*Endpoint, error) {
	start := time.Now()
	cfg := Config{}
	err := ut.MapToStruct(ut.TagNameNone, gwCfg.Provider.Config, &cfg)
	if err != nil {
		return nil, err
	}

	// endpoint
	d := &Endpoint{
		GwCfg:          gwCfg,
		Config:         cfg,
		receiveMsgFunc: rxMsgFunc,
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
	topics := rawMsg.Others.Get(gwptcl.KeyTopic).([]string)
	qos := byte(d.Config.QoS)
	for _, t := range topics {
		token := d.Client.Publish(t, qos, false, rawMsg.Data)
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
	return nil
}

func (d *Endpoint) getCallBack() func(paho.Client, paho.Message) {
	return func(c paho.Client, message paho.Message) {
		rawMsg := &msgml.RawMessage{
			Data:      message.Payload(),
			Timestamp: time.Now(),
			Others: map[string]interface{}{
				gwptcl.KeyTopic: message.Topic(),
				gwptcl.KeyQoS:   int(message.Qos()),
			},
		}
		err := d.receiveMsgFunc(rawMsg)
		if err != nil {
			zap.L().Error("Failed to send message to queue", zap.String("gateway", d.GwCfg.Name), zap.Any("rawMessage", rawMsg), zap.Error(err))
		}
	}
}

// Subscribe a topic
func (d *Endpoint) Subscribe(topic string) error {
	token := d.Client.Subscribe(topic, 0, d.getCallBack())
	token.WaitTimeout(3 * time.Second)
	return token.Error()
}

package mqtt

import (
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	msg "github.com/mycontroller-org/mycontroller-v2/pkg/model/message"
	ut "github.com/mycontroller-org/mycontroller-v2/pkg/util"

	"go.uber.org/zap"
)

// Endpoint data
type Endpoint struct {
	GwCfg          *ml.GatewayConfig
	Config         ml.GatewayConfigMQTT
	receiveMsgFunc func(rm *msg.RawMessage) error
	Client         paho.Client
}

// New mqtt driver
func New(gwCfg *ml.GatewayConfig, rxMsgFunc func(rm *msg.RawMessage) error) (*Endpoint, error) {
	start := time.Now()
	cfg := ml.GatewayConfigMQTT{}
	err := ut.MapToStruct(ut.TagNameNone, gwCfg.Provider.Config, &cfg)
	if err != nil {
		return nil, err
	}

	opts := paho.NewClientOptions()
	opts.AddBroker(cfg.Broker)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)
	opts.SetClientID(ut.RandID())
	opts.SetCleanSession(false)
	opts.SetAutoReconnect(true)
	opts.SetOnConnectHandler(onConnectionHandler)
	opts.SetConnectionLostHandler(onConnectionLostHandler)

	c := paho.NewClient(opts)
	token := c.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		return nil, err
	}

	d := &Endpoint{
		GwCfg:          gwCfg,
		Config:         cfg,
		Client:         c,
		receiveMsgFunc: rxMsgFunc,
	}

	err = d.Subscribe(cfg.Subscribe)
	if err != nil {
		zap.L().Error("Failed to subscribe a topic", zap.String("topic", cfg.Subscribe), zap.Error(err))
	}
	zap.L().Debug("MQTT client connected successfully", zap.String("timeTaken", time.Since(start).String()), zap.Any("clientConfig", cfg))
	return d, nil
}

func onConnectionHandler(c paho.Client) {
	zap.L().Debug("MQTT connection success")
}

func onConnectionLostHandler(c paho.Client, err error) {
	zap.L().Error("MQTT connection lost", zap.Error(err))
}

// Write publishes a payload
func (d *Endpoint) Write(rm *msg.RawMessage) error {
	topics := rm.Others[msg.KeyTopic].([]string)
	qos := byte(d.Config.QoS)
	for _, t := range topics {
		token := d.Client.Publish(t, qos, false, rm.Data)
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
		zap.L().Debug("MQTT Client connection closed", zap.Any("endpoint", d.Config))
	}
	return nil
}

func (d *Endpoint) getCallBack() func(paho.Client, paho.Message) {
	return func(c paho.Client, message paho.Message) {
		m := &msg.RawMessage{
			Data:      message.Payload(),
			Timestamp: time.Now(),
			Others: map[string]interface{}{
				msg.KeyTopic: message.Topic(),
				msg.KeyQoS:   int(message.Qos()),
			},
		}
		err := d.receiveMsgFunc(m)
		if err != nil {
			zap.L().Error("Failed to send message to queue", zap.Any("rawMessage", m), zap.Error(err))
		}
	}
}

// Subscribe a topic
func (d *Endpoint) Subscribe(topic string) error {
	token := d.Client.Subscribe(topic, 0, d.getCallBack())
	token.WaitTimeout(3 * time.Second)
	return token.Error()
}

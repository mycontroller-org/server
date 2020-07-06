package mqtt

import (
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	q "github.com/jaegertracing/jaeger/pkg/queue"
	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	msg "github.com/mycontroller-org/mycontroller-v2/pkg/model/message"
	ut "github.com/mycontroller-org/mycontroller-v2/pkg/util"

	"go.uber.org/zap"
)

// Endpoint data
type Endpoint struct {
	Config    ml.GatewayConfigMQTT
	Client    paho.Client
	RxQueue   *q.BoundedQueue
	TxQueue   *q.BoundedQueue
	GatewayID string
}

// New mqtt driver
func New(config map[string]interface{}, txQueue, rxQueue *q.BoundedQueue, gID string) (*Endpoint, error) {
	start := time.Now()
	cfg := ml.GatewayConfigMQTT{}
	err := ut.MapToStruct(ut.TagNameNone, config, &cfg)
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
		Config:    cfg,
		Client:    c,
		TxQueue:   txQueue,
		RxQueue:   rxQueue,
		GatewayID: gID,
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
	d.TxQueue.Stop()
	d.RxQueue.Stop()
	return nil
}

func (d *Endpoint) getCallBack() func(paho.Client, paho.Message) {
	return func(c paho.Client, message paho.Message) {
		m := &msg.Wrapper{
			GatewayID:  d.GatewayID,
			IsReceived: true,
			Message: &msg.RawMessage{
				Data:      message.Payload(),
				Timestamp: time.Now(),
				Others: map[string]interface{}{
					msg.KeyTopic: message.Topic(),
					msg.KeyQoS:   int(message.Qos()),
				},
			},
		}
		d.RxQueue.Produce(m)
		//	zap.L().Info("Received a message", zap.String("topic", msg.Topic()), zap.String("payload", string(msg.Payload())))
	}
}

// Subscribe a topic
func (d *Endpoint) Subscribe(topic string) error {
	token := d.Client.Subscribe(topic, 0, d.getCallBack())
	token.WaitTimeout(3 * time.Second)
	return token.Error()
}

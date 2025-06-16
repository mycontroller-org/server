package mqtt

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

const (
	PluginMqtt = "mqtt"

	timeout               = time.Second * 10
	loggerName            = "handler_mqtt"
	reconnectDelayDefault = time.Second * 30 // 30 seconds
)

// MqttConfig for mqtt
type MqttConfig struct {
	ClientID       string
	Broker         string
	Username       string
	Password       string `json:"-" yaml:"-"` // ignore password on logger
	Publish        string
	QoS            int
	Insecure       bool
	ReconnectDelay string
}

// MqttHandler struct
type MqttHandler struct {
	ID         string
	HandlerCfg *handlerTY.Config
	Config     *MqttConfig
	mqttClient paho.Client
	logger     *zap.Logger
}

func New(ctx context.Context, handlerCfg *handlerTY.Config) (handlerTY.Plugin, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	config := &MqttConfig{}
	err = utils.MapToStruct(utils.TagNameNone, handlerCfg.Spec, config)
	if err != nil {
		return nil, err
	}
	namedLogger := logger.Named(loggerName)

	namedLogger.Debug("mqtt client", zap.String("ID", handlerCfg.ID), zap.Any("config", config))

	client := &MqttHandler{
		ID:         handlerCfg.ID,
		HandlerCfg: handlerCfg,
		Config:     config,
		logger:     namedLogger,
	}

	// generate client id
	if config.ClientID == "" {
		config.ClientID = fmt.Sprintf("myc-handler-%s", utils.RandIDWithLength(5))
	}

	opts := paho.NewClientOptions()
	opts.AddBroker(config.Broker)
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)
	opts.SetClientID(config.ClientID)
	opts.SetCleanSession(false)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetryInterval(utils.ToDuration(config.ReconnectDelay, reconnectDelayDefault))
	opts.SetOnConnectHandler(client.onConnectionHandler)
	opts.SetConnectionLostHandler(client.onConnectionLostHandler)

	// update tls config
	tlsConfig := &tls.Config{InsecureSkipVerify: config.Insecure}
	opts.SetTLSConfig(tlsConfig)

	// adding client
	mqttClient := paho.NewClient(opts)

	client.mqttClient = mqttClient

	return client, nil
}

func (p *MqttHandler) Name() string {
	return PluginMqtt
}

// Start handler implementation
func (c *MqttHandler) Start() error {
	if c.mqttClient == nil {
		return errors.New("mqttClient can not be empty")
	}

	if c.mqttClient.IsConnected() {
		return nil
	}

	start := time.Now()

	token := c.mqttClient.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		return err
	}

	c.logger.Debug("mqtt client handler connected successfully", zap.String("timeTaken", time.Since(start).String()), zap.String("handlerId", c.ID), zap.String("clientId", c.Config.ClientID), zap.Any("clientConfig", c.Config))

	return nil
}

// Close handler implementation
func (c *MqttHandler) Close() error {
	if c.mqttClient != nil && c.mqttClient.IsConnected() {
		c.mqttClient.Disconnect(uint(time.Second * 5))
	}
	c.mqttClient = nil
	return nil

}

// State implementation
func (c *MqttHandler) State() *types.State {
	if c.HandlerCfg != nil {
		if c.HandlerCfg.State == nil {
			c.HandlerCfg.State = &types.State{}
		}
		return c.HandlerCfg.State
	}
	return &types.State{}
}

// Post handler implementation
func (c *MqttHandler) Post(parameters map[string]interface{}) error {
	if c.mqttClient == nil || !c.mqttClient.IsConnected() {
		return fmt.Errorf("mqtt client is not available to post a message, handler:%s, client Id:%s, broker:%s", c.ID, c.Config.ClientID, c.Config.Broker)
	}

	for name, rawParameter := range parameters {
		parameter, ok := handlerTY.IsTypeOf(rawParameter, handlerTY.DataTypeMqtt)
		if !ok {
			continue
		}
		c.logger.Debug("data", zap.String("name", name), zap.Any("parameter", parameter))

		mqttData := handlerTY.MqttData{}
		err := utils.MapToStruct(utils.TagNameNone, parameter, &mqttData)
		if err != nil {
			c.logger.Error("error on converting mqtt data", zap.Error(err), zap.String("name", name), zap.Any("parameter", parameter))
			continue
		}

		// replace defaults with custom values
		topic := c.Config.Publish
		if mqttData.Publish != "" {
			topic = mqttData.Publish
		}

		if topic == "" {
			return fmt.Errorf("topic can not be empty, name:%s, handlerId:%s", name, c.ID)
		}

		qos := c.Config.QoS
		if mqttData.QoS != -1 {
			qos = mqttData.QoS
		}

		// send data
		token := c.mqttClient.Publish(topic, byte(qos), false, mqttData.Data)
		if token.Error() != nil {
			return token.Error()
		}
	}

	return nil
}

func (c *MqttHandler) onConnectionHandler(pahoClient paho.Client) {
	c.logger.Debug("mqtt connection success", zap.String("clientId", c.Config.ClientID))
	c.HandlerCfg.State = &types.State{
		Status:  types.StatusUp,
		Message: "Connected successfully",
		Since:   time.Now(),
	}
}

func (c *MqttHandler) onConnectionLostHandler(pahoClient paho.Client, err error) {
	c.logger.Error("mqtt connection lost", zap.String("clientId", c.Config.ClientID), zap.Error(err))
	c.HandlerCfg.State = &types.State{
		Status:  types.StatusDown,
		Message: err.Error(),
		Since:   time.Now(),
	}
}

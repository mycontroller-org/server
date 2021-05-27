package natsio

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/json"
	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	busPlugin "github.com/mycontroller-org/backend/v2/plugin/bus"
	nats "github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// Config details of the client
type Config struct {
	Type                  string            `yaml:"type"`
	ServerURL             string            `yaml:"server_url"`
	TLSInsecureSkipVerify bool              `yaml:"tls_insecure_skip_verify"`
	ConnectionTimeout     string            `yaml:"connection_timeout"`
	ReconnectBufferSize   int               `yaml:"reconnect_buffer_size"`
	MaximumReconnect      int               `yaml:"maximum_reconnect"`
	ReconnectWait         string            `yaml:"reconnect_wait"`
	WebsocketOptions      *WebsocketOptions `yaml:"websocket_options"`
}

// WebsocketOptions are config options for a websocket dialer
type WebsocketOptions struct {
	RequestHeader   http.Header `yaml:"request_header"`
	ReadBufferSize  int         `yaml:"read_buffer_size"`
	WriteBufferSize int         `yaml:"write_buffer_size"`
}

// Client struct
type Client struct {
	ctx                 context.Context
	natConn             *nats.Conn
	topics              map[string][]int64
	subscriptions       map[int64]*nats.Subscription
	subscriptionCounter int64
	mutex               sync.RWMutex
	config              *Config
}

const (
	reconnectWaitDefault       = 5 * time.Second
	connectionTimeoutDefault   = 10 * time.Second
	maximumReconnectDefault    = 100
	reconnectBufferSizeDefault = 1000
)

// Init nats.io client
func Init(config cmap.CustomMap) (busPlugin.Client, error) {
	cfg := &Config{}
	err := utils.MapToStruct(utils.TagNameYaml, config, cfg)
	if err != nil {
		return nil, err
	}

	// set default values, if non set
	if cfg.ReconnectBufferSize == 0 {
		cfg.ReconnectBufferSize = reconnectBufferSizeDefault
	}
	if cfg.MaximumReconnect == 0 {
		cfg.MaximumReconnect = maximumReconnectDefault
	}
	if cfg.ServerURL == "" {
		cfg.ServerURL = nats.DefaultURL
	}

	// we handle tls with our custom dialer
	// say we are using "nats" protocol to nats.io client
	fakeServerURI, err := url.Parse(cfg.ServerURL)
	if err != nil {
		return nil, err
	}
	fakeServerURI.Scheme = "nats"

	opts := nats.Options{
		Url:               fakeServerURI.String(),
		Secure:            false, // will be handled by our custom dialer
		Verbose:           true,
		ReconnectWait:     utils.ToDuration(cfg.ReconnectWait, reconnectWaitDefault),
		AllowReconnect:    true,
		MaxReconnect:      cfg.MaximumReconnect,
		ReconnectBufSize:  cfg.ReconnectBufferSize,
		ClosedCB:          cbClosed,
		ReconnectedCB:     cbReconnected,
		DisconnectedCB:    cbDisconnected,
		DisconnectedErrCB: cbDisconnectedError,
	}

	customDialer, err := NewCustomDialer(cfg)
	if err != nil {
		return nil, err
	}
	opts.CustomDialer = customDialer
	nc, err := opts.Connect()
	if err != nil {
		return nil, err
	}
	client := Client{
		ctx:                 context.TODO(),
		natConn:             nc,
		topics:              make(map[string][]int64),
		subscriptions:       make(map[int64]*nats.Subscription),
		subscriptionCounter: 0,
		config:              cfg,
	}
	return &client, nil
}

// Close implementation
func (c *Client) Close() error {
	if c.natConn != nil {
		c.natConn.Close()
	}
	return nil
}

// Publish a data to a topic
func (c *Client) Publish(topic string, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	PrintDebug("Posting message", zap.String("topic", topic))
	return c.natConn.Publish(topic, bytes)
}

// Subscribe a topic
func (c *Client) Subscribe(topic string, handler busPlugin.CallBackFunc) (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	subscriptionIDs, found := c.topics[topic]
	if !found {
		subscriptionIDs = make([]int64, 0)
	}

	newSubscriptionID := c.generateSubscriptionID()
	wrappedHandler := c.handlerWrapper(handler)
	subscription, err := c.natConn.Subscribe(topic, wrappedHandler)
	if err != nil {
		return -1, err
	}
	c.subscriptions[newSubscriptionID] = subscription
	subscriptionIDs = append(subscriptionIDs, newSubscriptionID)
	c.topics[topic] = subscriptionIDs
	PrintDebug("Subscription created", zap.String("topic", subscription.Subject), zap.Int64("subscriptionID", newSubscriptionID))
	return newSubscriptionID, nil
}

func (c *Client) handlerWrapper(handler busPlugin.CallBackFunc) func(natsMsg *nats.Msg) {
	return func(natsMsg *nats.Msg) {
		PrintDebug("Receiving message", zap.String("topic", natsMsg.Sub.Subject))
		handler(&busML.BusData{Topic: natsMsg.Subject, Data: natsMsg.Data})
	}
}

// Unsubscribe a topic
func (c *Client) Unsubscribe(topic string, subscriptionID int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var subscription *nats.Subscription
	// remove subscription id
	if subscriptionIDs, found := c.topics[topic]; found {
		for index, id := range subscriptionIDs {
			if id == subscriptionID {
				subscription = c.subscriptions[id]
				c.topics[topic] = append(subscriptionIDs[:index], subscriptionIDs[index+1:]...)
				PrintDebug("Subscription removed", zap.String("topic", topic), zap.Int64("subscriptionID", subscriptionID))
				break
			}
		}
	}

	// remove subscription reference
	delete(c.subscriptions, subscriptionID)
	if subscription != nil {
		return subscription.Unsubscribe()
	}

	return nil
}

// UnsubscribeAll topics
func (c *Client) UnsubscribeAll(topic string) error {
	return errors.New("not implemented")
}

// call back functions
func cbDisconnected(con *nats.Conn) {
	PrintDebug("disconnected")
}

func cbReconnected(con *nats.Conn) {
	PrintDebug("reconnected")
}

func cbClosed(con *nats.Conn) {
	PrintDebug("connection closed")
}

func cbDisconnectedError(con *nats.Conn, err error) {
	PrintWarn("disconnected", zap.String("error", err.Error()))
}

func (c *Client) generateSubscriptionID() int64 {
	// increment counter id
	c.subscriptionCounter++
	return c.subscriptionCounter
}

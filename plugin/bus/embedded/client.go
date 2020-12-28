package embedded

import (
	"context"
	"strings"

	mbus "github.com/mustafaturan/bus"
	busml "github.com/mycontroller-org/backend/v2/plugin/bus"

	"github.com/mustafaturan/monoton"
	"github.com/mustafaturan/monoton/sequencer"
	"go.uber.org/zap"
)

// Client struct
type Client struct {
	ctx       context.Context
	busObject *mbus.Bus
}

// Init func
func Init() (busml.Client, error) {
	node := uint64(1)
	initialTime := uint64(1577865600000) // set 2020-01-01 PST as initial time
	m, err := monoton.New(sequencer.NewMillisecond(), node, initialTime)
	if err != nil {
		return nil, err
	}
	// init an id generator
	var idGenerator mbus.Next = (*m).Next
	// create a new bus instance
	b, err := mbus.NewBus(idGenerator)
	if err != nil {
		return nil, err
	}
	client := &Client{
		ctx:       context.TODO(),
		busObject: b,
	}
	return client, nil
}

// Close implementation
func (c *Client) Close() error {
	if c.busObject != nil {
		// deregister handlers
		for _, hk := range c.busObject.HandlerKeys() {
			c.busObject.DeregisterHandler(hk)
		}
		// deregister topics
		for _, t := range c.busObject.Topics() {
			c.busObject.DeregisterTopics(t)
		}
	}
	return nil
}

// Publish a data to a topic
func (c *Client) Publish(topic string, data interface{}) error {
	_, err := c.busObject.Emit(c.ctx, topic, data)
	if err != nil && strings.Contains(err.Error(), "not found") {
		zap.L().Debug("[BUS:EMBEDDED] Topic not registered. Registering now", zap.String("topic", topic), zap.Any("data", data))
		// register a topic
		c.busObject.RegisterTopics(topic)
		_, err = c.busObject.Emit(c.ctx, topic, data)
		return err
	}
	return err
}

// Subscribe a topic
func (c *Client) Subscribe(topic string, handler func(event *busml.Event)) error {
	wrappedHandler := &mbus.Handler{
		Matcher: topic,
		Handle:  c.handlerWrapper(handler),
	}
	c.busObject.RegisterHandler(topic, wrappedHandler)
	return nil
}

func (c *Client) handlerWrapper(handler func(event *busml.Event)) func(mbusEvent *mbus.Event) {
	return func(mbusEvent *mbus.Event) {
		handler(&busml.Event{Data: mbusEvent.Data})
	}
}

// Unsubscribe a topic
func (c *Client) Unsubscribe(topic string) error {
	c.busObject.DeregisterHandler(topic)
	return nil
}

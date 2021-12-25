package embedded

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	busML "github.com/mycontroller-org/server/v2/pkg/model/bus"
	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	busType "github.com/mycontroller-org/server/v2/plugin/bus/type"
	"go.uber.org/zap"
)

const PluginEmbedded = "embedded"

// Client struct
type Client struct {
	topics              map[string][]int64
	subscriptions       map[int64]busType.CallBackFunc
	subscriptionCounter int64
	mutex               *sync.RWMutex
}

// NewClient func
func NewClient(config cmap.CustomMap) (busType.Plugin, error) {
	client := &Client{
		topics:              make(map[string][]int64),
		subscriptions:       make(map[int64]busType.CallBackFunc),
		subscriptionCounter: 0,
		mutex:               &sync.RWMutex{},
	}
	return client, nil
}

func (c *Client) Name() string {
	return PluginEmbedded
}

// Close implementation
func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// clear all the call backs and topics
	c.subscriptionCounter = 0
	c.topics = make(map[string][]int64)
	c.subscriptions = make(map[int64]busType.CallBackFunc)
	return nil
}

// Publish a data to a topic
func (c *Client) Publish(topic string, data interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	PrintDebug("Posting message", zap.String("topic", topic))

	for subscriptionTopic := range c.topics {
		subscriptionIDs := c.topics[subscriptionTopic]
		match, err := regexp.MatchString(subscriptionTopic, topic)
		if err != nil {
			zap.L().Error("error on matching topic", zap.String("publishTopic", topic), zap.String("subscriptionTopic", subscriptionTopic), zap.Error(err))
			continue
		}

		if match {
			for _, subscriptionID := range subscriptionIDs {
				if callBack, ok := c.subscriptions[subscriptionID]; ok {
					event := &busML.BusData{Topic: topic}
					err := event.SetData(data)
					if err != nil {
						zap.L().Error("data conversion failed", zap.Error(err))
					} else {
						go callBack(event)
					}
				}
			}
		}
	}

	return nil
}

// Subscribe a topic
func (c *Client) Subscribe(topic string, handler busType.CallBackFunc) (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	topic = c.getFormatedTopic(topic)

	subscriptionIDs, found := c.topics[topic]
	if !found {
		subscriptionIDs = make([]int64, 0)
	}

	newSubscriptionID := c.generateSubscriptionID()
	c.subscriptions[newSubscriptionID] = handler
	subscriptionIDs = append(subscriptionIDs, newSubscriptionID)
	c.topics[topic] = subscriptionIDs
	PrintDebug("Subscription created", zap.String("topic", topic), zap.Int64("subscriptionID", newSubscriptionID))

	return newSubscriptionID, nil
}

// QueueSubscribe not supported in embedded bus, just call subscribe
func (c *Client) QueueSubscribe(topic, _queueName string, handler busType.CallBackFunc) (int64, error) {
	return c.Subscribe(topic, handler)
}

// QueueUnsubscribe not supported in embedded bus, just call subscribe
func (c *Client) QueueUnsubscribe(topic, _queueName string, subscriptionID int64) error {
	return c.Unsubscribe(topic, subscriptionID)
}

// Unsubscribe a topic
func (c *Client) Unsubscribe(topic string, subscriptionID int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	topic = c.getFormatedTopic(topic)

	// remove subscription id
	if subscriptionIDs, found := c.topics[topic]; found {
		for index, id := range subscriptionIDs {
			if id == subscriptionID {
				c.topics[topic] = append(subscriptionIDs[:index], subscriptionIDs[index+1:]...)
				PrintDebug("Subscription removed", zap.String("topic", topic), zap.Int64("subscriptionID", subscriptionID))
				break
			}
		}
	}

	// remove call back
	delete(c.subscriptions, subscriptionID)

	return nil
}

// UnsubscribeAll topics
func (c *Client) UnsubscribeAll(topic string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// remove subscription id
	if subscriptionIDs, found := c.topics[topic]; found {
		for _, subscriptionID := range subscriptionIDs {
			// remove call back
			delete(c.subscriptions, subscriptionID)
		}
		// delete topic
		delete(c.topics, topic)
	}

	return nil
}

func (c *Client) generateSubscriptionID() int64 {
	// increment counter id
	c.subscriptionCounter++

	return c.subscriptionCounter
}

func (c *Client) getFormatedTopic(topic string) string {
	updatedTopic := topic
	if strings.HasSuffix(topic, ">") {
		updatedTopic = fmt.Sprintf("%s*", topic[:len(topic)-1])
	}
	updatedTopic = fmt.Sprintf("^%s", updatedTopic)
	return updatedTopic
}

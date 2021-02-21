package embedded

import (
	"sync"

	"github.com/mycontroller-org/backend/v2/pkg/model/event"
	busml "github.com/mycontroller-org/backend/v2/plugin/bus"
	"go.uber.org/zap"
)

// Client struct
type Client struct {
	topics              map[string][]int64
	subscriptions       map[int64]busml.CallBackFunc
	subscriptionCounter int64
	mutex               sync.RWMutex
}

// Init func
func Init() (busml.Client, error) {
	client := &Client{
		topics:              make(map[string][]int64),
		subscriptions:       make(map[int64]busml.CallBackFunc),
		subscriptionCounter: 0,
	}
	return client, nil
}

// Close implementation
func (c *Client) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// clear all the call backs and topics
	c.subscriptionCounter = 0
	c.topics = make(map[string][]int64)
	c.subscriptions = make(map[int64]busml.CallBackFunc)
	return nil
}

// Publish a data to a topic
func (c *Client) Publish(topic string, data interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	PrintDebug("Posting message", zap.String("topic", topic))

	if subscriptionIDs, found := c.topics[topic]; found {
		for _, subscriptionID := range subscriptionIDs {
			if callBack, ok := c.subscriptions[subscriptionID]; ok {
				event := &event.Event{Topic: topic}
				err := event.SetData(data)
				if err != nil {
					zap.L().Error("data conversion failed", zap.Error(err))
				} else {
					go callBack(event)
				}
			}
		}
	}

	return nil
}

// Subscribe a topic
func (c *Client) Subscribe(topic string, handler busml.CallBackFunc) (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

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

// Unsubscribe a topic
func (c *Client) Unsubscribe(topic string, subscriptionID int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

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

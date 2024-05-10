package embedded

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/concurrency"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	busPluginTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

const (
	PluginEmbedded = "embedded"
	loggerName     = "BUS:EMBEDDED"
)

// Config details of the client
type Config struct {
	Type        string `yaml:"type"`
	TopicPrefix string `yaml:"topic_prefix"`
}

// Client struct
type Client struct {
	topics              map[string][]int64
	subscriptions       map[int64]busPluginTY.CallBackFunc
	subscriptionCounter int64
	mutex               *sync.RWMutex
	pauseFlag           concurrency.SafeBool
	logger              *zap.Logger
	config              *Config
}

// NewClient func
func NewClient(ctx context.Context, config cmap.CustomMap) (busPluginTY.Plugin, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = utils.MapToStruct(utils.TagNameYaml, config, cfg)
	if err != nil {
		return nil, err
	}

	client := &Client{
		topics:              make(map[string][]int64),
		subscriptions:       make(map[int64]busPluginTY.CallBackFunc),
		subscriptionCounter: 0,
		mutex:               &sync.RWMutex{},
		pauseFlag:           concurrency.SafeBool{},
		logger:              logger.Named(loggerName),
		config:              cfg,
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
	c.subscriptions = make(map[int64]busPluginTY.CallBackFunc)
	return nil
}

func (c *Client) TopicPrefix() string {
	return c.config.TopicPrefix
}

func (c *Client) PausePublish() {
	c.pauseFlag.Set()
}

func (c *Client) ResumePublish() {
	c.pauseFlag.Reset()
}

// Publish a data to a topic
func (c *Client) Publish(topic string, data interface{}) error {
	if c.pauseFlag.IsSet() {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// format topic with prefix
	topic = c.formatTopic(topic)

	c.logger.Debug("Posting message", zap.String("topic", topic))

	for subscriptionTopic := range c.topics {
		subscriptionIDs := c.topics[subscriptionTopic]
		match, err := regexp.MatchString(subscriptionTopic, topic)
		if err != nil {
			c.logger.Error("error on matching topic", zap.String("publishTopic", topic), zap.String("subscriptionTopic", subscriptionTopic), zap.Error(err))
			continue
		}

		if match {
			for _, subscriptionID := range subscriptionIDs {
				if callBack, ok := c.subscriptions[subscriptionID]; ok {
					event := &busTY.BusData{Topic: topic}
					err := event.SetData(data)
					if err != nil {
						c.logger.Error("data conversion failed", zap.Error(err))
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
func (c *Client) Subscribe(topic string, handler busPluginTY.CallBackFunc) (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// format topic with prefix
	topic = c.formatTopic(topic)

	topic = c.formatTopicWideSubscription(topic)

	subscriptionIDs, found := c.topics[topic]
	if !found {
		subscriptionIDs = make([]int64, 0)
	}

	newSubscriptionID := c.generateSubscriptionID()
	c.subscriptions[newSubscriptionID] = handler
	subscriptionIDs = append(subscriptionIDs, newSubscriptionID)
	c.topics[topic] = subscriptionIDs
	c.logger.Debug("Subscription created", zap.String("topic", topic), zap.Int64("subscriptionID", newSubscriptionID))

	return newSubscriptionID, nil
}

// QueueSubscribe not supported in embedded bus, just call subscribe
func (c *Client) QueueSubscribe(topic, _queueName string, handler busPluginTY.CallBackFunc) (int64, error) {
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

	// format topic with prefix
	topic = c.formatTopic(topic)

	// format topic with wide subscription
	topic = c.formatTopicWideSubscription(topic)

	// remove subscription id
	if subscriptionIDs, found := c.topics[topic]; found {
		for index, id := range subscriptionIDs {
			if id == subscriptionID {
				c.topics[topic] = append(subscriptionIDs[:index], subscriptionIDs[index+1:]...)
				c.logger.Debug("Subscription removed", zap.String("topic", topic), zap.Int64("subscriptionID", subscriptionID))
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

		// format topic with prefix
		topic = c.formatTopic(topic)

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

func (c *Client) formatTopic(topic string) string {
	if c.config.TopicPrefix != "" {
		return fmt.Sprintf("%s.%s", c.config.TopicPrefix, topic)
	}
	return topic
}

func (c *Client) formatTopicWideSubscription(topic string) string {
	updatedTopic := topic
	if strings.HasSuffix(topic, ">") {
		updatedTopic = fmt.Sprintf("%s*", topic[:len(topic)-1])
	}
	updatedTopic = fmt.Sprintf("^%s", updatedTopic)
	return updatedTopic
}

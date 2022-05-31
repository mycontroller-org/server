package provider

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/types"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	cloneUtils "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	gwPlugin "github.com/mycontroller-org/server/v2/plugin/gateway"
	providerTY "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

const (
	messageQueueLimit         = 200
	rawMessageQueueLimit      = 200
	sleepingQueuePerNodeLimit = 20 // number of messages can be in sleeping queue per node
	messageWorkersCount       = 1
	rawMessageWorkersCount    = 1

	defaultReconnectDelay = "15s"
)

// Service component of the provider
type Service struct {
	GatewayConfig                     *gwTY.Config
	provider                          providerTY.Plugin
	messageQueue                      *queueUtils.Queue
	rawMessageQueue                   *queueUtils.Queue
	topicPostToProvider               string
	topicPostToMsgProcessor           string
	topicPostToProviderSubscriptionID int64
	sleepingMessageQueue              map[string][]msgTY.Message
	mutex                             *sync.RWMutex
	ctx                               context.Context
}

// GetService returns service instance
func GetService(gatewayCfg *gwTY.Config) (*Service, error) {
	// verify default reconnect delay
	gatewayCfg.ReconnectDelay = utils.ValidDuration(gatewayCfg.ReconnectDelay, defaultReconnectDelay)

	// get a plugin
	provider, err := gwPlugin.Create(gatewayCfg.Provider.GetString(types.KeyType), gatewayCfg)
	if err != nil {
		return nil, err
	}
	service := &Service{
		GatewayConfig: gatewayCfg,
		provider:      provider,
		ctx:           context.TODO(),
		mutex:         &sync.RWMutex{},
	}
	return service, nil
}

// Start gateway service
func (s *Service) Start() error {
	zap.L().Debug("starting a provider service", zap.String("gatewayId", s.GatewayConfig.ID), zap.String("providerType", s.GatewayConfig.Provider.GetString(types.NameType)))

	// update topics
	s.topicPostToProvider = mcbus.GetTopicPostMessageToProvider(s.GatewayConfig.ID)
	s.topicPostToMsgProcessor = mcbus.GetTopicPostMessageToProcessor()

	s.messageQueue = queueUtils.New(fmt.Sprintf("gateway_provider_message_%s", s.GatewayConfig.ID), messageQueueLimit, s.messageConsumer, messageWorkersCount)
	s.rawMessageQueue = queueUtils.New(fmt.Sprintf("gateway_provider_raw_message_%s", s.GatewayConfig.ID), rawMessageQueueLimit, s.rawMessageProcessor, rawMessageWorkersCount)

	s.sleepingMessageQueue = make(map[string][]msgTY.Message)

	// start message listener
	s.startMessageListener()

	// start provider
	err := s.provider.Start(s.rawMessageQueueProduceFunc)
	if err != nil {
		zap.L().Error("error", zap.Error(err))
		err := mcbus.Unsubscribe(s.topicPostToProvider, s.topicPostToProviderSubscriptionID)
		if err != nil {
			zap.L().Error("error on unsubscribe", zap.Error(err), zap.String("topic", s.topicPostToProvider))
		}

		s.messageQueue.Close()
		s.rawMessageQueue.Close()
	}
	return err
}

// unsubscribe and stop queues
func (s *Service) stopService() {
	err := mcbus.Unsubscribe(s.topicPostToProvider, s.topicPostToProviderSubscriptionID)
	if err != nil {
		zap.L().Error("error on unsubscribe", zap.Error(err), zap.String("topic", s.topicPostToProvider))
	}
	s.messageQueue.Close()
	s.rawMessageQueue.Close()
}

// Stop the service
func (s *Service) Stop() error {
	defer s.stopService()     // in any case when exit, call stopService
	err := s.provider.Close() // close protocol connection
	if err != nil {
		return err
	}
	return nil
}

// this function supplied to provider-protocol
// rawMessages will be added directly to here
func (s *Service) rawMessageQueueProduceFunc(rawMsg *msgTY.RawMessage) error {
	status := s.rawMessageQueue.Produce(rawMsg)
	if !status {
		return errors.New("failed to add rawMessage in to queue")
	}
	return nil
}

// listens messages from server
func (s *Service) startMessageListener() {
	subscriptionID, err := mcbus.Subscribe(s.topicPostToProvider, func(event *busTY.BusData) {
		msg := &msgTY.Message{}
		err := event.LoadData(msg)
		if err != nil {
			zap.L().Warn("received invalid type", zap.Any("event", event))
			return
		}
		if msg.GatewayID == "" {
			zap.L().Warn("received message with empty gatewayId", zap.Any("message", msg))
			return
		}
		s.messageQueue.Produce(msg)
	})

	if err != nil {
		zap.L().Error("error on subscription", zap.String("topic", s.topicPostToProvider), zap.Error(err))
	} else {
		s.topicPostToProviderSubscriptionID = subscriptionID
	}
}

func (s *Service) messageConsumer(item interface{}) {
	msg, ok := item.(*msgTY.Message)
	if !ok {
		zap.L().Error("invalid message type", zap.Any("received", item))
		return
	}

	// include timestamp, if not set
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// if it is awake message send the sleeping queue messages
	if msg.Type == msgTY.TypeAction && len(msg.Payloads) > 0 && msg.Payloads[0].Value == nodeTY.ActionAwake {
		s.publishSleepingMessageQueue(msg.NodeID)
		return
	} else if msg.IsSleepNode {
		// for sleeping node message to be added in to the sleeping queue
		// when node sends awake signal, queue message will be published
		s.addToSleepingMessageQueue(msg)
		return
	} else {
		s.postMessage(msg, s.GatewayConfig.QueueFailedMessage)
	}
}

// postMessage to the provider
func (s *Service) postMessage(msg *msgTY.Message, queueFailedMessage bool) {
	err := s.provider.Post(msg)
	if err != nil {
		zap.L().Warn("error on sending", zap.String("gatewayId", s.GatewayConfig.ID), zap.Any("message", msg), zap.Error(err))
		if queueFailedMessage {
			msg.Labels = msg.Labels.Init()
			isSleepingQueueDisabled := msg.Labels.IsExists(types.LabelNodeSleepQueueDisabled) && msg.Labels.GetBool(types.LabelNodeSleepQueueDisabled)
			if !isSleepingQueueDisabled {
				s.addToSleepingMessageQueue(msg)
			}
		}
	}
}

// process received raw messages from protocol
func (s *Service) rawMessageProcessor(data interface{}) {
	rawMsg := data.(*msgTY.RawMessage)
	zap.L().Debug("received rawMessage", zap.String("gatewayId", s.GatewayConfig.ID), zap.Any("rawMessage", rawMsg))
	messages, err := s.provider.ConvertToMessages(rawMsg)
	if err != nil {
		zap.L().Warn("failed to parse", zap.String("gatewayId", s.GatewayConfig.ID), zap.Any("rawMessage", rawMsg), zap.Error(err))
		return
	}
	if len(messages) == 0 {
		zap.L().Debug("messages not parsed", zap.String("gatewayId", s.GatewayConfig.ID), zap.Any("rawMessage", rawMsg))
		return
	}
	// update gatewayID if not found
	for index := 0; index < len(messages); index++ {
		msg := messages[index]
		if msg != nil {
			if msg != nil && msg.GatewayID == "" {
				msg.GatewayID = s.GatewayConfig.ID
			}
			err = mcbus.Publish(s.topicPostToMsgProcessor, msg)
			if err != nil {
				zap.L().Debug("failed to post on topic", zap.String("topic", s.topicPostToMsgProcessor), zap.String("gatewayId", s.GatewayConfig.ID), zap.Any("message", msg), zap.Error(err))
				return
			}
		}
	}
}

// add message in to sleeping queue
func (s *Service) addToSleepingMessageQueue(msg *msgTY.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// add into sleeping queue
	queue, ok := s.sleepingMessageQueue[msg.NodeID]
	if !ok {
		queue = make([]msgTY.Message, 0)
		s.sleepingMessageQueue[msg.NodeID] = queue
	}
	// verify if the message already in the queue, if then remove it
	newMsgId := msg.GetID()
	for index, oldMsg := range queue {
		if newMsgId == oldMsg.GetID() {
			queue[index] = *msg
			queue = append(queue[:index], queue[index+1:]...)
			break
		}
	}

	// add it to the queue
	queue = append(queue, *msg)

	// if queue size exceeds maximum defined size, do resize
	oldSize := len(queue)
	if oldSize > sleepingQueuePerNodeLimit {
		queue = queue[:sleepingQueuePerNodeLimit]
		zap.L().Debug("dropped messages from sleeping queue", zap.Int("oldSize", oldSize), zap.Int("newSize", len(queue)), zap.String("gatewayID", msg.GatewayID), zap.String("nodeID", msg.NodeID))
	}
	s.sleepingMessageQueue[msg.NodeID] = queue
}

// emptySleepingMessageQueue clears all the messages on the queue
func (s *Service) publishSleepingMessageQueue(nodeID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	msgQueue, ok := s.sleepingMessageQueue[nodeID]
	if !ok {
		return
	}
	// post messages
	for _, msg := range msgQueue {
		s.postMessage(&msg, false)
	}

	// remove messages from the map
	s.sleepingMessageQueue[nodeID] = make([]msgTY.Message, 0)
}

// returns sleeping queue messages
func (s *Service) GetGatewaySleepingQueue() map[string][]msgTY.Message {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// clone the queue and return
	clonedQueue := cloneUtils.Clone(s.sleepingMessageQueue)
	return clonedQueue.(map[string][]msgTY.Message)
}

// returns sleeping queue messages
func (s *Service) GetNodeSleepingQueue(nodeID string) []msgTY.Message {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	msgQueue, ok := s.sleepingMessageQueue[nodeID]
	if !ok {
		return make([]msgTY.Message, 0)
	}
	// clone the queue and return
	clonedQueue := cloneUtils.Clone(msgQueue)
	return clonedQueue.([]msgTY.Message)
}

// clear all sleeping messages
func (s *Service) ClearGatewaySleepingQueue() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// clear all the messages
	s.sleepingMessageQueue = make(map[string][]msgTY.Message)
}

// clear sleeping queue message of a node
func (s *Service) ClearNodeSleepingQueue(nodeID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// remove messages for a node
	s.sleepingMessageQueue[nodeID] = make([]msgTY.Message, 0)
}

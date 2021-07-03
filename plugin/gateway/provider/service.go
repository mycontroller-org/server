package provider

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/mycontroller-org/server/v2/pkg/model"
	busML "github.com/mycontroller-org/server/v2/pkg/model/bus"
	msgML "github.com/mycontroller-org/server/v2/pkg/model/message"
	nodeML "github.com/mycontroller-org/server/v2/pkg/model/node"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	gwPlugin "github.com/mycontroller-org/server/v2/plugin/gateway"
	providerType "github.com/mycontroller-org/server/v2/plugin/gateway/provider/type"
	gwType "github.com/mycontroller-org/server/v2/plugin/gateway/type"
	"go.uber.org/zap"
)

const (
	queueSizeMessage          = 200
	queueSizeRawMessage       = 200
	sleepingQueuePerNodeLimit = 20
	workersMessage            = 1
	workersRawMessage         = 1

	defaultReconnectDelay = "15s"
)

// Service component of the provider
type Service struct {
	GatewayConfig                     *gwType.Config
	provider                          providerType.Plugin
	messageQueue                      *queueUtils.Queue
	rawMessageQueue                   *queueUtils.Queue
	topicListenFromServer             string
	topicPostToServer                 string
	topicListenFromCoreSubscriptionID int64
	sleepingMessageQueue              map[string][]msgML.Message
	mutex                             *sync.RWMutex
	ctx                               context.Context
}

// GetService returns service instance
func GetService(gatewayCfg *gwType.Config) (*Service, error) {
	// verify default reconnect delay
	gatewayCfg.ReconnectDelay = utils.ValidDuration(gatewayCfg.ReconnectDelay, defaultReconnectDelay)

	provider, err := gwPlugin.Create(gatewayCfg.Provider.GetString(model.KeyType), gatewayCfg)
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
	zap.L().Debug("starting a provider service", zap.String("gateway", s.GatewayConfig.ID), zap.String("provider", s.GatewayConfig.Provider.GetString(model.NameType)))

	// update topics
	s.topicListenFromServer = mcbus.GetTopicPostMessageToProvider(s.GatewayConfig.ID)
	s.topicPostToServer = mcbus.GetTopicPostMessageToServer()

	s.messageQueue = queueUtils.New(fmt.Sprintf("gateway_provider_message_%s", s.GatewayConfig.ID), queueSizeMessage, s.messageConsumer, workersMessage)
	s.rawMessageQueue = queueUtils.New(fmt.Sprintf("gateway_provider_raw_message_%s", s.GatewayConfig.ID), queueSizeRawMessage, s.rawMessageProcessor, workersRawMessage)

	s.sleepingMessageQueue = make(map[string][]msgML.Message)

	// start message listener
	s.startMessageListener()

	// start provider
	err := s.provider.Start(s.addRawMessageToQueueFunc)
	if err != nil {
		zap.L().Error("error", zap.Error(err))
		err := mcbus.Unsubscribe(s.topicListenFromServer, s.topicListenFromCoreSubscriptionID)
		if err != nil {
			zap.L().Error("error on unsubscribing a topic", zap.Error(err), zap.String("topic", s.topicListenFromServer))
		}

		s.messageQueue.Close()
		s.rawMessageQueue.Close()
	}
	return err
}

// unsubscribe and stop queues
func (s *Service) stopService() {
	err := mcbus.Unsubscribe(s.topicListenFromServer, s.topicListenFromCoreSubscriptionID)
	if err != nil {
		zap.L().Error("error on unsubscribe", zap.Error(err), zap.String("topic", s.topicListenFromServer))
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

// this function supplied to protocol
// rawMessages will be added directely here
func (s *Service) addRawMessageToQueueFunc(rawMsg *msgML.RawMessage) error {
	status := s.rawMessageQueue.Produce(rawMsg)
	if !status {
		return errors.New("failed to add rawmessage in to queue")
	}
	return nil
}

// listens messages from server
func (s *Service) startMessageListener() {
	subscriptionID, err := mcbus.Subscribe(s.topicListenFromServer, func(event *busML.BusData) {
		msg := &msgML.Message{}
		err := event.LoadData(msg)
		if err != nil {
			zap.L().Warn("received invalid type", zap.Any("event", event))
			return
		}
		if msg == nil {
			zap.L().Warn("received a nil message")
			return
		}
		s.messageQueue.Produce(msg)
	})

	if err != nil {
		zap.L().Error("error on subscription", zap.String("topic", s.topicListenFromServer), zap.Error(err))
	} else {
		s.topicListenFromCoreSubscriptionID = subscriptionID
	}
}

func (s *Service) messageConsumer(item interface{}) {
	msg, ok := item.(*msgML.Message)
	if !ok {
		zap.L().Error("invalid message type", zap.Any("received", item))
		return
	}

	// if it is awake message send the sleeping queue messages
	if msg.Type == msgML.TypeAction && len(msg.Payloads) > 0 && msg.Payloads[0].Key == nodeML.ActionAwake {
		s.publishSleepingMessageQueue(msg.NodeID)
		return
	} else if msg.IsPassiveNode {
		// for sleeping node message put it on sleeping queue and exit
		s.addToSleepingMessageQueue(msg)
		return
	} else {
		s.postMessage(msg, true)
	}

}

// postMessage to the provider
func (s *Service) postMessage(msg *msgML.Message, addToSleepingQueue bool) {
	err := s.provider.Post(msg)
	if err != nil {
		zap.L().Warn("error on sending", zap.String("gateway", s.GatewayConfig.ID), zap.Any("message", msg), zap.Error(err))
		if addToSleepingQueue && s.GatewayConfig.QueueFailedMessage {
			s.addToSleepingMessageQueue(msg)
		}
	}
}

// process received raw messages from protocol
func (s *Service) rawMessageProcessor(data interface{}) {
	rawMsg := data.(*msgML.RawMessage)
	zap.L().Debug("rawMessage received", zap.String("gateway", s.GatewayConfig.ID), zap.Any("rawMessage", rawMsg))
	messages, err := s.provider.ProcessReceived(rawMsg)
	if err != nil {
		zap.L().Warn("failed to parse", zap.String("gateway", s.GatewayConfig.ID), zap.Any("rawMessage", rawMsg), zap.Error(err))
		return
	}
	if len(messages) == 0 {
		zap.L().Debug("messages not parsed", zap.String("gateway", s.GatewayConfig.ID), zap.Any("rawMessage", rawMsg))
		return
	}
	// update gatewayID if not found
	for index := 0; index < len(messages); index++ {
		msg := messages[index]
		if msg != nil {
			if msg != nil && msg.GatewayID == "" {
				msg.GatewayID = s.GatewayConfig.ID
			}
			err = mcbus.Publish(s.topicPostToServer, msg)
			if err != nil {
				zap.L().Debug("messages failed to post on topic", zap.String("topic", s.topicPostToServer), zap.String("gateway", s.GatewayConfig.ID), zap.Any("message", msg), zap.Error(err))
				return
			}
		}
	}
}

// add message in to sleeping queue
func (s *Service) addToSleepingMessageQueue(msg *msgML.Message) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// add into sleeping queue
	queue, ok := s.sleepingMessageQueue[msg.NodeID]
	if !ok {
		queue = make([]msgML.Message, 0)
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
		zap.L().Debug("dropped messags from sleeping queue", zap.Int("oldSize", oldSize), zap.Int("newSize", len(queue)), zap.String("gatewayID", msg.GatewayID), zap.String("nodeID", msg.NodeID))
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

	// remove mesages from the map
	s.sleepingMessageQueue[nodeID] = make([]msgML.Message, 0)
}

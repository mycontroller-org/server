package gateway

import (
	"fmt"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/mustafaturan/bus"
	"github.com/mycontroller-org/backend/v2/pkg/mcbus"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"go.uber.org/zap"
)

const (
	messageToProviderQueueLimit = 1000 // number of messages can in the queue, will be in memory
	messageToProviderWorkers    = 1    // number of parallel workers, this should be always 1 for most of the case
	sleepingMsgQueueLimit       = 50   // number of sleeping messages per node can be hold
)

func getQueue(name string, queueSize int) *q.BoundedQueue {
	return q.NewBoundedQueue(queueSize, func(data interface{}) {
		zap.L().Error("Queue full. Droping a data on gateway side", zap.String("name", name), zap.Any("message", data))
	})
}

// Start gateway service
func (s *Service) Start() error {
	zap.L().Debug("Starting a gateway service", zap.String("name", s.GatewayConfig.Name))

	// update topics
	s.topicMessageToProvider = fmt.Sprintf("%s_%s", mcbus.TopicMessageToGateway, s.GatewayConfig.ID)
	s.topicSleepingMessageToProvider = fmt.Sprintf("%s_%s", mcbus.TopicSleepingMessageToGateway, s.GatewayConfig.ID)

	s.messageToProviderQueue = getQueue("txQueue", messageToProviderQueueLimit)
	s.sleepingNodeMessageQueue = make(map[string][]*msgml.Message)

	// start handlers
	s.startMessageToProviderHandler()
	s.startSleepingMessageToProviderHandler()

	// start provider
	err := s.provider.Start(s.messageFromProviderHandlerFunc())
	if err != nil {
		zap.L().Error("Error", zap.Error(err))
		mcbus.BUS.DeregisterTopics(s.topicMessageToProvider)
		mcbus.BUS.DeregisterTopics(s.topicSleepingMessageToProvider)
		s.messageToProviderQueue.Stop()
	}
	return err
}

// Stop the media
func (s *Service) Stop() error {
	mcbus.BUS.DeregisterTopics(s.topicMessageToProvider)
	mcbus.BUS.DeregisterTopics(s.topicSleepingMessageToProvider)
	s.messageToProviderQueue.Stop()
	// stop media
	err := s.provider.Close()
	if err != nil {
		zap.L().Debug("Error", zap.Error(err))
	}
	return nil
}

// handler to send messages to gateway
func (s *Service) startMessageToProviderHandler() {
	mcbus.Subscribe(s.topicMessageToProvider, &bus.Handler{
		Matcher: s.topicMessageToProvider,
		Handle: func(e *bus.Event) {
			s.messageToProviderQueue.Produce(e.Data)
		},
	})
	s.messageToProviderQueue.StartConsumers(messageToProviderWorkers, func(data interface{}) {
		msg := data.(*msgml.Message)
		if msg.IsPassiveNode {
			s.addMessageInToSleepingQueue(msg, sleepingMsgQueueLimit)
		} else {
			s.postMessageToProvider(msg)
		}
	})
}

func (s *Service) startSleepingMessageToProviderHandler() {
	// subscribe sleeping message
	mcbus.Subscribe(s.topicSleepingMessageToProvider, &bus.Handler{
		Matcher: s.topicSleepingMessageToProvider,
		Handle: func(e *bus.Event) {
			msg, ok := e.Data.(*msgml.Message)
			if !ok {
				zap.L().Error("received invalid data type", zap.Any("event", e))
				return
			}
			queueMessages := s.GetSleepingMessageQueue(msg.NodeID)
			for _, queueMsg := range queueMessages {
				s.postMessageToProvider(queueMsg)
			}
		},
	})

	// consume messages on the queue and send it provider
	s.messageToProviderQueue.StartConsumers(messageToProviderWorkers, func(data interface{}) {
		msg := data.(*msgml.Message)
		if msg.IsPassiveNode {
			// add into sleeping queue
			queue, ok := s.sleepingNodeMessageQueue[msg.NodeID]
			if !ok {
				queue = make([]*msgml.Message, 0)
				s.sleepingNodeMessageQueue[msg.NodeID] = queue
			}
			queue = append(queue, msg)
			// if queue size exceeds maximum defined size, do resize
			oldSize := len(queue)
			if oldSize > sleepingMsgQueueLimit {
				queue = queue[:sleepingMsgQueueLimit]
				zap.L().Debug("Dropped messags from sleeping queue", zap.Int("oldSize", oldSize), zap.Int("newSize", len(queue)))
			}
		} else {
			s.postMessageToProvider(msg)
		}
	})
}

// AddMessageToSleepingQueue into queue
func (s *Service) addMessageInToSleepingQueue(msg *msgml.Message, limit int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// add into sleeping queue
	queue, ok := s.sleepingNodeMessageQueue[msg.NodeID]
	if !ok {
		queue = make([]*msgml.Message, 0)
		s.sleepingNodeMessageQueue[msg.NodeID] = queue
	}
	queue = append(queue, msg)
	// if queue size exceeds maximum defined size, do resize
	oldSize := len(queue)
	if oldSize > limit {
		queue = queue[:limit]
		zap.L().Debug("Dropped messags from sleeping queue", zap.Int("oldSize", oldSize), zap.Int("newSize", len(queue)))
	}
}

// ClearSleepingMessageQueue clears all the messages on the queue
func (s *Service) ClearSleepingMessageQueue() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.sleepingNodeMessageQueue = make(map[string][]*msgml.Message)
}

// GetSleepingMessageQueue returns message for a specific nodeID, also removes it from the queue
func (s *Service) GetSleepingMessageQueue(nodeID string) []*msgml.Message {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if queue, ok := s.sleepingNodeMessageQueue[nodeID]; ok {
		s.sleepingNodeMessageQueue[nodeID] = make([]*msgml.Message, 0)
		return queue
	}
	return nil
}

func (s *Service) postMessageToProvider(msg *msgml.Message) {
	// update gatewayID if not found
	if msg != nil && msg.GatewayID == "" {
		msg.GatewayID = s.GatewayConfig.ID
	}

	// parse message
	rawMsg, err := s.provider.ToRawMessage(msg)
	if err != nil {
		postMessageDeliveryStatus(&msgml.DeliveryStatus{ID: msg.GetID(), Success: false, Message: err.Error()})
		zap.L().Error("Message failed to parse", zap.Any("message", msg), zap.Error(err))
		return
	}

	// post message
	err = s.provider.Post(rawMsg)
	if err != nil {
		zap.L().Error("Failed to post message", zap.Any("message", msg), zap.Any("rawMessage", rawMsg), zap.Error(err))
		postMessageDeliveryStatus(&msgml.DeliveryStatus{ID: msg.GetID(), Success: false, Message: err.Error()})
		return
	}
	postMessageDeliveryStatus(&msgml.DeliveryStatus{ID: msg.GetID(), Success: true, Message: "message sent successfully"})
	return
}

// post message delivery status
func postMessageDeliveryStatus(status *msgml.DeliveryStatus) {
	_, err := mcbus.Publish(mcbus.TopicMessageToGatewayDelieverStatus, status)
	if err != nil {
		zap.L().Error("Failed to send delivery status", zap.Error(err))
	}
}

// handle messages from provider
func (s *Service) messageFromProviderHandlerFunc() func(rawMsg *msgml.RawMessage) error {
	return func(rawMsg *msgml.RawMessage) error {
		zap.L().Debug("RawMessage received", zap.String("gateway", s.GatewayConfig.Name), zap.Any("rawMessage", rawMsg))
		messages, err := s.provider.ToMessage(rawMsg)
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			zap.L().Debug("Messages not parsed", zap.String("gateway", s.GatewayConfig.Name), zap.Any("RawMessage", rawMsg))
			return nil
		}
		// update gatewayID if not found
		for index := 0; index < len(messages); index++ {
			msg := messages[index]
			if msg != nil {
				if msg != nil && msg.GatewayID == "" {
					msg.GatewayID = s.GatewayConfig.ID
				}
				var topic string
				if msg.IsAck {
					topic = fmt.Sprintf("%s_%s", mcbus.TopicGatewayAcknowledgement, msg.GetID())
				} else {
					topic = mcbus.TopicMessageFromGateway
				}
				_, err = mcbus.Publish(topic, msg)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
}

package gateway

import (
	"fmt"
	"time"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/mustafaturan/bus"
	"github.com/mycontroller-org/backend/v2/pkg/mcbus"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"go.uber.org/zap"
)

const (
	txQueueLimit          = 1000 // number of messages can in the queue, will be on the RAM
	txQueueWorks          = 1    // this should be always 1 for most of the case
	sleepingMsgQueueLimit = 50   // number of messages per node can be hold
)

func getQueue(name string, queueSize int) *q.BoundedQueue {
	return q.NewBoundedQueue(queueSize, func(data interface{}) {
		zap.L().Error("Queue full. Droping a data on gateway side", zap.String("name", name), zap.Any("message", data))
	})
}

// handle received messages from gateway
func (s *Service) receiveMessageFunc() func(rawMsg *msgml.RawMessage) error {
	return func(rawMsg *msgml.RawMessage) error {
		zap.L().Debug("RawMessage received", zap.String("gateway", s.Config.Name), zap.Any("rawMessage", rawMsg))
		messages, err := s.Provider.ToMessage(rawMsg)
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			zap.L().Debug("Messages not parsed", zap.String("gateway", s.Config.Name), zap.Any("RawMessage", rawMsg))
			return nil
		}
		// update gatewayID if not found
		for index := 0; index < len(messages); index++ {
			msg := messages[index]
			if msg != nil {
				if msg != nil && msg.GatewayID == "" {
					msg.GatewayID = s.Config.ID
				}
				var topic string
				if msg.IsAck {
					topic = fmt.Sprintf("%s_%s", mcbus.TopicGatewayAcknowledgement, msg.GetID())
				} else {
					topic = mcbus.TopicMsgFromGW
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

func (s *Service) writeMessageFunc() func(msg *msgml.Message) {
	return func(msg *msgml.Message) {
		// update gatewayID if not found
		if msg != nil && msg.GatewayID == "" {
			msg.GatewayID = s.Config.ID
		}

		// send delivery status
		postDeliveryStatus := func(status *msgml.DeliveryStatus) {
			_, err := mcbus.Publish(mcbus.TopicMsg2GWDelieverStatus, status)
			if err != nil {
				zap.L().Error("Failed to send delivery status", zap.Error(err))
			}
		}

		// check acknowledgement status and enable if required
		// enable ack only for sensor fields
		// set to false default
		msg.IsAckEnabled = false
		// is ack enabled?, is node id available?, is active node?
		if s.Config.Ack.Enabled && msg.NodeID != "" && !msg.IsPassiveNode {
			msg.IsAckEnabled = true
		}

		// parse message
		rawMsg, err := s.Provider.ToRawMessage(msg)
		if err != nil {
			postDeliveryStatus(&msgml.DeliveryStatus{ID: msg.GetID(), Success: false, Message: err.Error()})
			zap.L().Error("Message failed to parse", zap.Any("message", msg), zap.Error(err))
			return
		}

		// write logic
		if msg.IsAckEnabled {
			// wait for acknowledgement message
			ackChannel := make(chan bool, 1)
			ackTopic := fmt.Sprintf("%s_%s", mcbus.TopicGatewayAcknowledgement, rawMsg.ID)
			mcbus.Subscribe(ackTopic, &bus.Handler{
				Matcher: ackTopic,
				Handle: func(e *bus.Event) {
					zap.L().Debug("channel close issue", zap.Any("event", e))
					// TODO: facing issue, closed channel. Fix this
					ackChannel <- true
				},
			})

			// on exit unsubscribe and close the channel
			defer mcbus.Unsubscribe(ackTopic)
			defer close(ackChannel)

			timeout, err := time.ParseDuration(s.Config.Ack.Timeout)
			if err != nil {
				// failed to parse timeout, running with default
				timeout = time.Millisecond * 200
				zap.L().Warn("Failed to parse timeout, running with default timeout", zap.String("timeout", s.Config.Ack.Timeout), zap.Error(err))
			}

			// minimum timeout
			if timeout.Microseconds() < 1 {
				timeout = time.Millisecond * 10
			}

			messageSent := false
			for retry := 1; retry <= s.Config.Ack.RetryCount; retry++ {
				// write into gateway
				err = s.Provider.Post(rawMsg)
				if err != nil {
					zap.L().Error("Failed to post message into provider gw", zap.Error(err), zap.Any("message", msg), zap.Any("rawMessage", rawMsg))
					postDeliveryStatus(&msgml.DeliveryStatus{ID: msg.GetID(), Success: false, Message: err.Error()})
					return
				}
				select {
				case <-ackChannel:
					messageSent = true
				case <-time.After(timeout):
					// just wait till timeout
				}
				if messageSent {
					break
				}
			}
			if messageSent {
				postDeliveryStatus(&msgml.DeliveryStatus{ID: msg.GetID(), Success: true})
			} else {
				postDeliveryStatus(&msgml.DeliveryStatus{ID: msg.GetID(), Success: false, Message: "No acknowledgement received, tried maximum retries"})
			}
			return
		}

		// message without acknowledgement
		err = s.Provider.Post(rawMsg)
		if err != nil {
			zap.L().Error("Failed to write message into gateway", zap.Error(err), zap.Any("message", msg), zap.Any("raw", rawMsg))
			postDeliveryStatus(&msgml.DeliveryStatus{ID: msg.GetID(), Success: false, Message: err.Error()})
			return
		}
		postDeliveryStatus(&msgml.DeliveryStatus{ID: msg.GetID(), Success: true})
	}
}

// handle message to gateway
func (s *Service) handleMessageToGateway(writeFunc func(msg *msgml.Message)) {
	mcbus.Subscribe(s.TopicMsg2Provider, &bus.Handler{
		Matcher: s.TopicMsg2Provider,
		Handle: func(e *bus.Event) {
			s.OutMsgQueue.Produce(e.Data)
		},
	})
	s.OutMsgQueue.StartConsumers(txQueueWorks, func(data interface{}) {
		msg := data.(*msgml.Message)
		if msg.IsPassiveNode {
			// disable ack for sleeping node messages
			msg.IsAckEnabled = false
			s.AddSleepMsg(msg, sleepingMsgQueueLimit)
		} else {
			writeFunc(msg)
		}
	})
}

func (s *Service) handleSleepingMessage(writeFunc func(msg *msgml.Message)) {
	handleSleepingmsgFunc := func(data interface{}) {
		msg := data.(*msgml.Message)
		queue := s.GetSleepingQueue(msg.NodeID)
		for _, _msg := range queue {
			writeFunc(_msg)
		}
	}

	mcbus.Subscribe(s.TopicSleepingMsg2Provider, &bus.Handler{
		Matcher: s.TopicSleepingMsg2Provider,
		Handle: func(e *bus.Event) {
			handleSleepingmsgFunc(e.Data)
		},
	})
	s.OutMsgQueue.StartConsumers(txQueueWorks, func(data interface{}) {
		msg := data.(*msgml.Message)
		if msg.IsPassiveNode {
			// add into sleeping queue
			queue, ok := s.SleepingNodeMsgQueue[msg.NodeID]
			if !ok {
				queue = make([]*msgml.Message, 0)
				s.SleepingNodeMsgQueue[msg.NodeID] = queue
			}
			queue = append(queue, msg)
			// if queue size exceeds maximum defined size, do resize
			oldSize := len(queue)
			if oldSize > sleepingMsgQueueLimit {
				queue = queue[:sleepingMsgQueueLimit]
				zap.L().Debug("Dropped messags from sleeping queue", zap.Int("oldSize", oldSize), zap.Int("newSize", len(queue)))
			}
		} else {
			writeFunc(msg)
		}
	})
}

// Start gateway service
func (s *Service) Start() error {
	zap.L().Debug("gateway config", zap.Any("gw", s.Config))

	s.TopicMsg2Provider = fmt.Sprintf("%s_%s", mcbus.TopicMsg2GW, s.Config.ID)
	s.TopicSleepingMsg2Provider = fmt.Sprintf("%s_%s", mcbus.TopicSleepingMsg2GW, s.Config.ID)
	s.OutMsgQueue = getQueue("txQueue", txQueueLimit)
	s.SleepingNodeMsgQueue = make(map[string][]*msgml.Message)

	txMessageFunc := s.writeMessageFunc()
	rxMessageFunc := s.receiveMessageFunc()

	s.handleMessageToGateway(txMessageFunc)
	s.handleSleepingMessage(txMessageFunc)

	err := s.Provider.Start(rxMessageFunc)
	if err != nil {
		zap.L().Error("Error", zap.Error(err))
		mcbus.BUS.DeregisterTopics(s.TopicMsg2Provider)
		mcbus.BUS.DeregisterTopics(s.TopicSleepingMsg2Provider)
		s.OutMsgQueue.Stop()
	}
	return err
}

// Stop the media
func (s *Service) Stop() error {
	mcbus.BUS.DeregisterTopics(s.TopicMsg2Provider)
	mcbus.BUS.DeregisterTopics(s.TopicSleepingMsg2Provider)
	s.OutMsgQueue.Stop()
	// stop media
	err := s.Provider.Close()
	if err != nil {
		zap.L().Debug("Error", zap.Error(err))
	}
	return nil
}

// AddSleepMsg into queue
func (s *Service) AddSleepMsg(msg *msgml.Message, limit int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// add into sleeping queue
	queue, ok := s.SleepingNodeMsgQueue[msg.NodeID]
	if !ok {
		queue = make([]*msgml.Message, 0)
		s.SleepingNodeMsgQueue[msg.NodeID] = queue
	}
	queue = append(queue, msg)
	// if queue size exceeds maximum defined size, do resize
	oldSize := len(queue)
	if oldSize > limit {
		queue = queue[:limit]
		zap.L().Debug("Dropped messags from sleeping queue", zap.Int("oldSize", oldSize), zap.Int("newSize", len(queue)))
	}
}

// ClearSleepingQueue clears all the messages on the queue
func (s *Service) ClearSleepingQueue() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.SleepingNodeMsgQueue = make(map[string][]*msgml.Message)
}

// GetSleepingQueue returns message for a specific nodeID, also removes it from the queue
func (s *Service) GetSleepingQueue(nodeID string) []*msgml.Message {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if queue, ok := s.SleepingNodeMsgQueue[nodeID]; ok {
		s.SleepingNodeMsgQueue[nodeID] = make([]*msgml.Message, 0)
		return queue
	}
	return nil
}

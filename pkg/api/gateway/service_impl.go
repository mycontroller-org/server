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
func (s *Service) receiveMessageFunc() func(rm *msgml.RawMessage) error {
	return func(rm *msgml.RawMessage) error {
		zap.L().Debug("rawMessage received", zap.Any("rawMessage", rm), zap.Any("payload", string(rm.Data)))
		message, err := s.Provider.ToMessage(rm)
		if err != nil {
			return err
		}
		if message == nil {
			zap.L().Debug("message not parsed", zap.Any("rawMessage", rm), zap.String("payload", string(rm.Data)))
			return nil
		}
		// update gatewayID if not found
		if message != nil && message.GatewayID == "" {
			message.GatewayID = s.Config.ID
		}
		var topic string
		if message.IsAck {
			topic = fmt.Sprintf("%s_%s", mcbus.TopicGatewayAcknowledgement, message.GetID())
		} else {
			topic = mcbus.TopicMsgFromGW
		}
		_, err = mcbus.Publish(topic, message)
		return err
	}
}

func (s *Service) writeMessageFunc() func(mcMsg *msgml.Message) {
	return func(mcMsg *msgml.Message) {
		// update gatewayID if not found
		if mcMsg != nil && mcMsg.GatewayID == "" {
			mcMsg.GatewayID = s.Config.ID
		}
		// send delivery status
		sendDeliveryStatus := func(status *msgml.DeliveryStatus) {
			_, err := mcbus.Publish(mcbus.TopicMsg2GWDelieverStatus, status)
			if err != nil {
				zap.L().Error("Failed to send delivery status", zap.Error(err))
			}
		}
		// parse message
		rm, err := s.Provider.ToRawMessage(mcMsg)
		if err != nil {
			sendDeliveryStatus(&msgml.DeliveryStatus{ID: mcMsg.GetID(), Success: false, Message: err.Error()})
			zap.L().Error("Failed to parse", zap.Error(err), zap.Any("message", mcMsg))
			return
		}

		// write logic
		if mcMsg.IsAckEnabled {
			// wait for acknowledgement message
			ackChannel := make(chan bool, 1)
			ackTopic := fmt.Sprintf("%s_%s", mcbus.TopicGatewayAcknowledgement, rm.ID)
			mcbus.Subscribe(ackTopic, &bus.Handler{
				Matcher: ackTopic,
				Handle: func(e *bus.Event) {
					ackChannel <- true
				},
			})
			defer mcbus.Unsubscribe(ackTopic)

			timeout, err := time.ParseDuration(s.Config.Ack.Timeout)
			if err != nil {
				// failed to parse timeout, update default
				timeout = time.Millisecond * 200
			}
			// minimum timeout
			if timeout.Microseconds() < 1 {
				timeout = time.Millisecond * 10
			}

			messageSent := false
			for retry := 1; retry <= s.Config.Ack.RetryCount; retry++ {
				// write into gateway
				err = s.Provider.Post(rm)
				if err != nil {
					zap.L().Error("Failed to write message into gateway", zap.Error(err), zap.Any("message", mcMsg), zap.Any("raw", rm))
					sendDeliveryStatus(&msgml.DeliveryStatus{ID: mcMsg.GetID(), Success: false, Message: err.Error()})
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
				sendDeliveryStatus(&msgml.DeliveryStatus{ID: mcMsg.GetID(), Success: true})
			} else {
				sendDeliveryStatus(&msgml.DeliveryStatus{ID: mcMsg.GetID(), Success: false, Message: "No acknowledgement received, tried maximum retries"})
			}
			return
		}

		// message without acknowledgement
		err = s.Provider.Post(rm)
		if err != nil {
			zap.L().Error("Failed to write message into gateway", zap.Error(err), zap.Any("message", mcMsg), zap.Any("raw", rm))
			sendDeliveryStatus(&msgml.DeliveryStatus{ID: mcMsg.GetID(), Success: false, Message: err.Error()})
			return
		}
		sendDeliveryStatus(&msgml.DeliveryStatus{ID: mcMsg.GetID(), Success: true})
	}
}

// handle message to gateway
func (s *Service) handleMessageToGateway(writeFunc func(mcMsg *msgml.Message)) {
	mcbus.Subscribe(s.TopicMsg2Provider, &bus.Handler{
		Matcher: s.TopicMsg2Provider,
		Handle: func(e *bus.Event) {
			s.OutMsgQueue.Produce(e.Data)
		},
	})
	s.OutMsgQueue.StartConsumers(txQueueWorks, func(data interface{}) {
		mcMsg := data.(*msgml.Message)
		if mcMsg.IsPassiveNode {
			// disable ack for sleeping node messages
			mcMsg.IsAckEnabled = false
			s.AddSleepMsg(mcMsg, sleepingMsgQueueLimit)
		} else {
			writeFunc(mcMsg)
		}
	})
}

func (s *Service) handleSleepingMessage(writeFunc func(mcMsg *msgml.Message)) {
	handleSleepingmsgFunc := func(data interface{}) {
		mcMsg := data.(*msgml.Message)
		queue := s.GetSleepingQueue(mcMsg.NodeID)
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
		mcMsg := data.(*msgml.Message)
		if mcMsg.IsPassiveNode {
			// add into sleeping queue
			queue, ok := s.SleepingNodeMsgQueue[mcMsg.NodeID]
			if !ok {
				queue = make([]*msgml.Message, 0)
				s.SleepingNodeMsgQueue[mcMsg.NodeID] = queue
			}
			queue = append(queue, mcMsg)
			// if queue size exceeds maximum defined size, do resize
			oldSize := len(queue)
			if oldSize > sleepingMsgQueueLimit {
				queue = queue[:sleepingMsgQueueLimit]
				zap.L().Debug("Dropped messags from sleeping queue", zap.Int("oldSize", oldSize), zap.Int("newSize", len(queue)))
			}
		} else {
			writeFunc(mcMsg)
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
	return nil
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
func (s *Service) AddSleepMsg(mcMsg *msgml.Message, limit int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	// add into sleeping queue
	queue, ok := s.SleepingNodeMsgQueue[mcMsg.NodeID]
	if !ok {
		queue = make([]*msgml.Message, 0)
		s.SleepingNodeMsgQueue[mcMsg.NodeID] = queue
	}
	queue = append(queue, mcMsg)
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

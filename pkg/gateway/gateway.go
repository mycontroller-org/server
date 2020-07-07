package gateway

import (
	"context"
	"fmt"
	"time"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/mustafaturan/bus"
	"github.com/mycontroller-org/mycontroller-v2/pkg/gateway/mqtt"
	"github.com/mycontroller-org/mycontroller-v2/pkg/gateway/serial"
	"github.com/mycontroller-org/mycontroller-v2/pkg/mcbus"
	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	msg "github.com/mycontroller-org/mycontroller-v2/pkg/model/message"
	srv "github.com/mycontroller-org/mycontroller-v2/pkg/service"
	"go.uber.org/zap"
)

const (
	txQueueLimit          = 1000 // number of messages can in the queue, will be on the RAM
	txQueueWorks          = 1    // this should be always 1 for most of the case
	sleepingMsgQueueLimit = 50   // number of messages per node can be hold
)

var ctx = context.TODO()

func getQueue(name string, queueSize int) *q.BoundedQueue {
	return q.NewBoundedQueue(queueSize, func(data interface{}) {
		zap.L().Error("Queue full. Droping a data on gateway side", zap.String("name", name), zap.Any("message", data))
	})
}

// handle received messages from gateway
func receiveMessageFunc(s *ml.GatewayService) func(rm *msg.RawMessage) error {
	return func(rm *msg.RawMessage) error {
		zap.L().Debug("rawMessage received", zap.Any("rawMessage", rm), zap.Any("payload", string(rm.Data)))
		message, err := s.Parser.ToMessage(rm)
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

func writeMessageFunc(s *ml.GatewayService) func(mcMsg *msg.Message) {
	return func(mcMsg *msg.Message) {
		// update gatewayID if not found
		if mcMsg != nil && mcMsg.GatewayID == "" {
			mcMsg.GatewayID = s.Config.ID
		}
		// send delivery status
		sendDeliveryStatus := func(status *msg.DeliveryStatus) {
			_, err := mcbus.Publish(mcbus.TopicMsg2GWDelieverStatus, status)
			if err != nil {
				zap.L().Error("Failed to send delivery status", zap.Error(err))
			}
		}
		// parse message
		rm, err := s.Parser.ToRawMessage(mcMsg)
		if err != nil {
			sendDeliveryStatus(&msg.DeliveryStatus{ID: mcMsg.GetID(), Success: false, Message: err.Error()})
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

			waitTime, err := time.ParseDuration(s.Config.AckConfig.WaitTime)
			if err != nil {
				// failed to parse waitTime, update default
				waitTime = time.Millisecond * 200
			}
			// minimum waittime
			if waitTime.Microseconds() < 1 {
				waitTime = time.Millisecond * 10
			}

			messageSent := false
			for retry := 1; retry <= s.Config.AckConfig.RetryCount; retry++ {
				// write into gateway
				err = s.Gateway.Write(rm)
				if err != nil {
					zap.L().Error("Failed to write message into gateway", zap.Error(err), zap.Any("message", mcMsg), zap.Any("raw", rm))
					sendDeliveryStatus(&msg.DeliveryStatus{ID: mcMsg.GetID(), Success: false, Message: err.Error()})
					return
				}
				select {
				case <-ackChannel:
					messageSent = true
				case <-time.After(waitTime):
					// breaks this wait
				}
				if messageSent {
					break
				}
			}
			if messageSent {
				sendDeliveryStatus(&msg.DeliveryStatus{ID: mcMsg.GetID(), Success: true})
			} else {
				sendDeliveryStatus(&msg.DeliveryStatus{ID: mcMsg.GetID(), Success: false, Message: "No acknowledgement received, tried maximum retries"})
			}
			return
		}

		// message without acknowledgement
		err = s.Gateway.Write(rm)
		if err != nil {
			zap.L().Error("Failed to write message into gateway", zap.Error(err), zap.Any("message", mcMsg), zap.Any("raw", rm))
			sendDeliveryStatus(&msg.DeliveryStatus{ID: mcMsg.GetID(), Success: false, Message: err.Error()})
			return
		}
		sendDeliveryStatus(&msg.DeliveryStatus{ID: mcMsg.GetID(), Success: true})
	}
}

// handle message to gateway
func handleMessageToGateway(s *ml.GatewayService, writeFunc func(mcMsg *msg.Message)) {
	mcbus.Subscribe(s.TopicMsg2GW, &bus.Handler{
		Matcher: s.TopicMsg2GW,
		Handle: func(e *bus.Event) {
			s.TxMsgQueue.Produce(e.Data)
		},
	})
	s.TxMsgQueue.StartConsumers(txQueueWorks, func(data interface{}) {
		mcMsg := data.(*msg.Message)
		if mcMsg.IsSleepingNode {
			// disable ack for sleeping node messages
			mcMsg.IsAckEnabled = false
			s.AddSleepMsg(mcMsg, sleepingMsgQueueLimit)
		} else {
			writeFunc(mcMsg)
		}
	})
}

func handleSleepingMessage(s *ml.GatewayService, writeFunc func(mcMsg *msg.Message)) {
	handleSleepingmsgFunc := func(data interface{}) {
		mcMsg := data.(*msg.Message)
		queue := s.GetSleepingQueue(mcMsg.NodeID)
		for _, _msg := range queue {
			writeFunc(_msg)
		}
	}

	mcbus.Subscribe(s.TopicSleepingMsg2GW, &bus.Handler{
		Matcher: s.TopicSleepingMsg2GW,
		Handle: func(e *bus.Event) {
			handleSleepingmsgFunc(e.Data)
		},
	})
	s.TxMsgQueue.StartConsumers(txQueueWorks, func(data interface{}) {
		mcMsg := data.(*msg.Message)
		if mcMsg.IsSleepingNode {
			// add into sleeping queue
			queue, ok := s.SleepMsgQueue[mcMsg.NodeID]
			if !ok {
				queue = make([]*msg.Message, 0)
				s.SleepMsgQueue[mcMsg.NodeID] = queue
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
func Start(s *ml.GatewayService) error {
	s.TopicMsg2GW = fmt.Sprintf("%s_%s", mcbus.TopicMsg2GW, s.Config.ID)
	s.TopicSleepingMsg2GW = fmt.Sprintf("%s_%s", mcbus.TopicSleepingMsg2GW, s.Config.ID)
	s.TxMsgQueue = getQueue("txQueue", txQueueLimit)
	s.SleepMsgQueue = make(map[string][]*msg.Message)

	txMessageFunc := writeMessageFunc(s)
	rxMessageFunc := receiveMessageFunc(s)

	handleMessageToGateway(s, txMessageFunc)
	handleSleepingMessage(s, txMessageFunc)

	var err error
	switch s.Config.Provider.GatewayType {
	case ml.GatewayTypeMQTT:
		ms, _err := mqtt.New(s.Config, rxMessageFunc)
		err = _err
		s.Gateway = ms
	case ml.GatewayTypeSerial:
		ms, _err := serial.New(s.Config, rxMessageFunc)
		err = _err
		s.Gateway = ms
	}
	if err != nil {
		zap.L().Error("Error", zap.Error(err))
		srv.BUS.DeregisterTopics(s.TopicMsg2GW)
		srv.BUS.DeregisterTopics(s.TopicSleepingMsg2GW)
		s.TxMsgQueue.Stop()
	}
	return nil
}

// Stop the media
func Stop(s *ml.GatewayService) error {
	srv.BUS.DeregisterTopics(s.TopicMsg2GW)
	srv.BUS.DeregisterTopics(s.TopicSleepingMsg2GW)
	s.TxMsgQueue.Stop()
	// stop media
	err := s.Gateway.Close()
	if err != nil {
		zap.L().Debug("Error", zap.Error(err))
	}
	return nil
}

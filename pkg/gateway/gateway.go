package gateway

import (
	"context"
	"fmt"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/mustafaturan/bus"
	"github.com/mycontroller-org/mycontroller/pkg/gateway/mqtt"
	"github.com/mycontroller-org/mycontroller/pkg/gateway/serial"
	"github.com/mycontroller-org/mycontroller/pkg/mcbus"
	ml "github.com/mycontroller-org/mycontroller/pkg/model"
	msg "github.com/mycontroller-org/mycontroller/pkg/model/message"
	srv "github.com/mycontroller-org/mycontroller/pkg/service"
	"go.uber.org/zap"
)

const (
	rxQueueSize  = 1000
	txQueueSize  = 100
	msgRxWorkers = 1
	msgTxWorkers = 1 // this should be always 1
)

var ctx = context.TODO()

func getQueue(name string, queueSize int) *q.BoundedQueue {
	return q.NewBoundedQueue(queueSize, func(data interface{}) {
		zap.L().Error("Queue full. Droping a data on gateway side", zap.String("name", name), zap.Any("message", data))
	})
}

// handle received messages from gateway
func handleRxMessages(s *ml.GatewayService, q *q.BoundedQueue) {
	topicRx := fmt.Sprintf("%s_%s", mcbus.TopicGatewayMessageRx, s.Config.ID)
	topicAckRx := fmt.Sprintf("%s_%s", mcbus.TopicGatewayMessageAckRx, s.Config.ID)
	q.StartConsumers(msgRxWorkers, func(data interface{}) {
		// zap.L().Debug("Raw message received", zap.Any("message", data))
		//start := time.Now()
		wm := data.(*msg.Wrapper)
		m, err := s.Parser.ToMessage(wm)
		if err != nil {
			zap.L().Error("Failed to parse", zap.Error(err), zap.Any("raw", wm))
		}
		// if message not nil post it on main bus
		if m != nil {
			m.GatewayID = wm.GatewayID
			// check is it ack message?
			if m.IsAck {
				_, err := mcbus.Publish(topicAckRx, m)
				if err != nil {
					zap.L().Error("Failed to post ack message to global bus", zap.Error(err), zap.Any("ackMessage", m))
				}
			} else {
				_, err := mcbus.Publish(topicRx, m)
				if err != nil {
					zap.L().Error("Failed to post message to global bus", zap.Error(err), zap.Any("message", m))
				}
			}
			// zap.L().Debug("Time taken", zap.Any("event", e), zap.String("timeTaken", time.Since(start).String()))
		}
	})
}

// handle tx messages
func handleTxMessages(s *ml.GatewayService, q *q.BoundedQueue) {
	topic := fmt.Sprintf("%s_%s", mcbus.TopicGatewayMessageTx, s.Config.ID)
	mcbus.Subscribe(topic, &bus.Handler{
		Matcher: topic,
		Handle: func(e *bus.Event) {
			q.Produce(e.Data)
		},
	})
	q.StartConsumers(msgTxWorkers, func(data interface{}) {
		//start := time.Now()
		wm := data.(*msg.Wrapper)
		mcMsg := wm.Message.(*msg.Message)
		rm, err := s.Parser.ToRawMessage(wm)
		if err != nil {
			zap.L().Error("Failed to parse", zap.Error(err), zap.Any("message", wm))
		}
		// send delivery status
		sendDeliveryStatus := func(status *msg.DeliveryStatus) {
			_, err = mcbus.Publish(mcbus.TopicGatewayMessageDelieverStatus, status)
			if err != nil {
				zap.L().Error("Failed to send delivery status", zap.Error(err))
			}
		}
		// write into gateway
		err = s.Gateway.Write(rm)
		if err != nil {
			zap.L().Error("Failed to write message into gateway", zap.Error(err), zap.Any("wm", wm), zap.Any("raw", rm))
			sendDeliveryStatus(&msg.DeliveryStatus{ID: mcMsg.ID, Success: false, Message: err.Error()})
			return
		}
		if s.Config.AckConfig.Enabled {
			// wait for ack
		} else {
			sendDeliveryStatus(&msg.DeliveryStatus{ID: mcMsg.ID, Success: true})
		}
	})
}

// Start gateway service
func Start(s *ml.GatewayService) error {
	txQueue := getQueue("txQueue", txQueueSize)
	rxQueue := getQueue("rxQueue", rxQueueSize)
	handleTxMessages(s, txQueue)
	handleRxMessages(s, rxQueue)

	var err error
	gwCfg := s.Config.Provider.Config
	switch s.Config.Provider.GatewayType {
	case ml.GatewayTypeMQTT:
		ms, _err := mqtt.New(gwCfg, txQueue, rxQueue, s.Config.ID)
		err = _err
		s.Gateway = ms
	case ml.GatewayTypeSerial:
		ms, _err := serial.New(gwCfg, txQueue, rxQueue, s.Config.ID)
		err = _err
		s.Gateway = ms
	}
	if err != nil {
		zap.L().Error("Error", zap.Error(err))
		txQueue.Stop()
		rxQueue.Stop()
	}
	return nil
}

// Stop the media
func Stop(s *ml.GatewayService) error {
	// de register this media topic from global bus
	srv.BUS.DeregisterTopics(s.Topics.PostMessage)
	srv.BUS.DeregisterTopics(s.Topics.PostAcknowledgement)

	// stop media
	err := s.Gateway.Close()
	if err != nil {
		zap.L().Debug("Error", zap.Error(err))
	}

	return nil
}

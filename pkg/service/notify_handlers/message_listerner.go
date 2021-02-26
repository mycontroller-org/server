package handler

import (
	"fmt"
	"time"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/event"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	"go.uber.org/zap"
)

var (
	msgQueue  *q.BoundedQueue
	queueSize = int(1000)
)

// init message listener
func initMessageListener() error {
	msgQueue = utils.GetQueue("handler_message_listener", queueSize)

	// on message receive add it in to our local queue
	_, err := mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicPostMessageNotifyHandler), onMessageReceive)
	if err != nil {
		return err
	}

	msgQueue.StartConsumers(1, processHandlerMessage)
	return nil
}

func onMessageReceive(event *event.Event) {
	msg := &handlerML.MessageWrapper{}
	err := event.ToStruct(msg)
	if err != nil {
		zap.L().Warn("Failed to convet to target type", zap.Error(err))
		return
	}

	if msg == nil {
		zap.L().Warn("Received a nil message", zap.Any("event", event))
		return
	}
	zap.L().Debug("Message added into processing queue", zap.Any("message", msg))
	status := msgQueue.Produce(msg)
	if !status {
		zap.L().Warn("Failed to store the message into queue", zap.Any("message", msg))
	}
}

// close listener service
func closeMessageListener() {
	msgQueue.Stop()
}

func processHandlerMessage(item interface{}) {
	msg := item.(*handlerML.MessageWrapper)
	start := time.Now()

	zap.L().Debug("Starting Message Processing", zap.Any("handlerID", msg.ID))

	handler := handlersStore.Get(msg.ID)
	if handler == nil {
		zap.L().Warn("handler not available", zap.Any("handlerID", msg.ID), zap.Any("availableHandlers", handlersStore.ListIDs()))
		return
	}

	state := handler.State()

	err := handler.Post(msg.Data)
	if err != nil {
		zap.L().Warn("failed to execute handler", zap.Any("handlerID", msg.ID), zap.Error(err))
		state.Status = model.StateError
		state.Message = err.Error()
	} else {
		state.Status = model.StateOk
		state.Message = fmt.Sprintf("execution time: %s", time.Since(start).String())
	}

	state.Since = time.Now()
	busUtils.SetHandlerState(msg.ID, *state)
}

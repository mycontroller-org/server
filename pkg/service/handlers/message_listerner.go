package handler

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

var (
	msgQueue  *queueUtils.Queue
	queueSize = int(1000)
	workers   = int(1)
)

// init message listener
func initMessageListener() error {
	msgQueue = queueUtils.New("handler_message_listener", queueSize, processHandlerMessage, workers)

	// on message receive add it in to our local queue
	_, err := mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicPostMessageNotifyHandler), onMessageReceive)
	if err != nil {
		return err
	}

	return nil
}

func onMessageReceive(event *busML.BusData) {
	msg := &handlerML.MessageWrapper{}
	err := event.LoadData(msg)
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
	msgQueue.Close()
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
		state.Status = model.StatusError
		state.Message = err.Error()
	} else {
		state.Status = model.StatusOk
		state.Message = fmt.Sprintf("execution time: %s", time.Since(start).String())
	}

	state.Since = time.Now()
	busUtils.SetHandlerState(msg.ID, *state)
}

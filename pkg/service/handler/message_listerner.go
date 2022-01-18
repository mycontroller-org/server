package handler

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
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

func onMessageReceive(event *busTY.BusData) {
	msg := &handlerTY.MessageWrapper{}
	err := event.LoadData(msg)
	if err != nil {
		zap.L().Warn("failed to convet to target type", zap.Error(err))
		return
	}

	if msg == nil {
		zap.L().Warn("received a nil message", zap.Any("event", event))
		return
	}
	zap.L().Debug("message added into processing queue", zap.Any("message", msg))
	status := msgQueue.Produce(msg)
	if !status {
		zap.L().Warn("failed to store the message into queue", zap.Any("message", msg))
	}
}

// close listener service
func closeMessageListener() {
	msgQueue.Close()
}

func processHandlerMessage(item interface{}) {
	msg := item.(*handlerTY.MessageWrapper)
	start := time.Now()

	zap.L().Debug("starting message processing", zap.Any("handlerID", msg.ID))

	handler := handlersStore.Get(msg.ID)
	if handler == nil {
		zap.L().Info("handler not available", zap.Any("handlerID", msg.ID), zap.Any("availableHandlers", handlersStore.ListIDs()))
		return
	}

	state := handler.State()

	err := handler.Post(msg.Data)
	if err != nil {
		zap.L().Warn("failed to execute handler", zap.Any("handlerID", msg.ID), zap.Error(err))
		state.Status = types.StatusError
		state.Message = err.Error()
	} else {
		state.Status = types.StatusOk
		state.Message = fmt.Sprintf("execution time: %s", time.Since(start).String())
	}

	state.Since = time.Now()
	busUtils.SetHandlerState(msg.ID, *state)
}

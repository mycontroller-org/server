package handler

import (
	"fmt"
	"time"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

// init message listener
func (svc *HandlerService) initMessageListener() error {
	// on message receive add it in to our local queue
	id, err := svc.bus.Subscribe(svc.messageQueue.Topic, svc.onMessageReceive)
	if err != nil {
		return err
	}
	svc.messageQueue.SubscriptionId = id

	return nil
}

func (svc *HandlerService) onMessageReceive(event *busTY.BusData) {
	msg := &handlerTY.MessageWrapper{}
	err := event.LoadData(msg)
	if err != nil {
		svc.logger.Warn("failed to convert to target type", zap.Error(err))
		return
	}

	if len(msg.Data) == 0 {
		svc.logger.Warn("received an empty event", zap.Any("event", event))
		return
	}
	svc.logger.Debug("message added into processing queue", zap.Any("message", msg))
	status := svc.messageQueue.Produce(msg)
	if !status {
		svc.logger.Warn("failed to store the message into queue", zap.Any("message", msg))
	}
}

// close listener service
func (svc *HandlerService) closeMessageListener() {
	svc.messageQueue.Close()
}

func (svc *HandlerService) processHandlerMessage(item interface{}) {
	msg := item.(*handlerTY.MessageWrapper)
	start := time.Now()

	svc.logger.Debug("starting message processing", zap.Any("handlerID", msg.ID))

	handler := svc.store.Get(msg.ID)
	if handler == nil {
		svc.logger.Info("handler not available", zap.Any("handlerID", msg.ID), zap.Any("availableHandlers", svc.store.ListIDs()))
		return
	}

	state := handler.State()

	err := handler.Post(msg.Data)
	if err != nil {
		svc.logger.Warn("failed to execute handler", zap.Any("handlerID", msg.ID), zap.Error(err))
		state.Status = types.StatusError
		state.Message = err.Error()
	} else {
		state.Status = types.StatusOk
		state.Message = fmt.Sprintf("execution time: %s", time.Since(start).String())
	}

	state.Since = time.Now()
	busUtils.SetHandlerState(svc.logger, svc.bus, msg.ID, *state)
}

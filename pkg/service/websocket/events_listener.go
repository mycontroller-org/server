package mcwebsocket

import (
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/mycontroller-org/server/v2/pkg/json"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	wsTY "github.com/mycontroller-org/server/v2/pkg/types/websocket"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

// starts events listener
func (svc *WebsocketService) startEventListener() error {
	// on message receive add it in to our local queue
	sID, err := svc.bus.Subscribe(svc.eventsQueue.Topic, svc.onEventReceive)
	if err != nil {
		return err
	}
	svc.eventsQueue.SubscriptionId = sID
	return nil
}

func (svc *WebsocketService) CloseEventListener() error {
	err := svc.bus.Unsubscribe(svc.eventsQueue.Topic, svc.eventsQueue.SubscriptionId)
	if err != nil {
		return err
	}
	svc.eventsQueue.Close()
	return nil
}

func (svc *WebsocketService) onEventReceive(data *busTY.BusData) {
	status := svc.eventsQueue.Produce(data)
	if !status {
		svc.logger.Error("failed to post a event on the processor queue")
	}
}

func (svc *WebsocketService) processEvent(item interface{}) error {
	// if there is no clients, just ignore the event
	if svc.store.getSize() == 0 {
		return nil
	}

	data := item.(*busTY.BusData)

	event := &eventTY.Event{}
	err := data.LoadData(event)
	if err != nil {
		svc.logger.Warn("failed to convert to target type", zap.Any("topic", data.Topic), zap.Error(err))
		return nil
	}

	svc.logger.Debug("event received", zap.Any("event", event))

	response := wsTY.Response{
		Type: wsTY.ResponseTypeEvent,
		Data: event,
	}

	// convert to json bytes
	dataBytes, err := json.Marshal(response)
	if err != nil {
		svc.logger.Error("error on converting to json", zap.Error(err))
		return nil
	}

	wsClients := svc.store.getClients()
	for index := range wsClients {
		client := wsClients[index]

		// write with write timeout
		err := client.SetWriteDeadline(time.Now().Add(defaultWriteTimeout))
		if err != nil {
			svc.logger.Debug("error on setting write deadline", zap.Any("remoteAddress", client.RemoteAddr().String()), zap.Error(err))
			svc.store.unregister(client)
			return nil
		}
		err = client.WriteMessage(ws.TextMessage, dataBytes)
		if err != nil {
			svc.logger.Debug("error on write data to a client", zap.Any("remoteAddress", client.RemoteAddr().String()), zap.Error(err))
			svc.store.unregister(client)
		}
	}
	return nil
}

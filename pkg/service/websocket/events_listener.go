package mcwebsocket

import (
	ws "github.com/gorilla/websocket"
	"github.com/mycontroller-org/server/v2/pkg/json"
	busML "github.com/mycontroller-org/server/v2/pkg/model/bus"
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
	wsML "github.com/mycontroller-org/server/v2/pkg/model/websocket"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

const (
	eventListenerQueueLimit  = 1000
	eventListenerWorkerLimit = 1
	eventListenerQueueName   = "websocket_event_listener"
)

var (
	eventsQueue          *queueUtils.Queue
	eventsSubscriptionID = int64(0)
	eventsTopic          = "" // updated dynamically
)

// initEventListener events listener
func initEventListener() error {
	eventsQueue = queueUtils.New(eventListenerQueueName, eventListenerQueueLimit, processEvent, eventListenerWorkerLimit)

	// on message receive add it in to our local queue
	eventsTopic = mcbus.FormatTopic(mcbus.TopicEventsAll)
	sID, err := mcbus.Subscribe(eventsTopic, onEventReceive)
	if err != nil {
		return err
	}
	eventsSubscriptionID = sID
	return nil
}

func CloseEventListener() error {
	err := mcbus.Unsubscribe(eventsTopic, eventsSubscriptionID)
	if err != nil {
		return err
	}
	eventsQueue.Close()
	return nil
}

func onEventReceive(data *busML.BusData) {
	status := eventsQueue.Produce(data)
	if !status {
		zap.L().Error("failed to post selected tasks on processor queue")
	}
}

func processEvent(item interface{}) {
	data := item.(*busML.BusData)

	event := &eventML.Event{}
	err := data.LoadData(event)
	if err != nil {
		zap.L().Warn("failed to convet to target type", zap.Any("topic", data.Topic), zap.Error(err))
		return
	}

	zap.L().Debug("event received", zap.Any("event", event))

	response := wsML.Response{
		Type: wsML.ResponseTypeEvent,
		Data: event,
	}

	// convert to json bytes
	dataBytes, err := json.Marshal(response)
	if err != nil {
		zap.L().Error("error on converting to json", zap.Error(err))
		return
	}

	for client := range clients {
		err := client.WriteMessage(ws.TextMessage, dataBytes)
		if err != nil {
			zap.L().Error("error on write to a client", zap.Error(err), zap.Any("client", client.LocalAddr().String()))
			client.Close()
			delete(clients, client)
		}
	}
}

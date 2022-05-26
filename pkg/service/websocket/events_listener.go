package mcwebsocket

import (
	ws "github.com/gorilla/websocket"
	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	wsTY "github.com/mycontroller-org/server/v2/pkg/types/websocket"
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

func onEventReceive(data *busTY.BusData) {
	status := eventsQueue.Produce(data)
	if !status {
		zap.L().Error("failed to post selected tasks on processor queue")
	}
}

func processEvent(item interface{}) {
	data := item.(*busTY.BusData)

	event := &eventTY.Event{}
	err := data.LoadData(event)
	if err != nil {
		zap.L().Warn("failed to convert to target type", zap.Any("topic", data.Topic), zap.Error(err))
		return
	}

	zap.L().Debug("event received", zap.Any("event", event))

	response := wsTY.Response{
		Type: wsTY.ResponseTypeEvent,
		Data: event,
	}

	// convert to json bytes
	dataBytes, err := json.Marshal(response)
	if err != nil {
		zap.L().Error("error on converting to json", zap.Error(err))
		return
	}

	wsClients := getClients()
	for index := range wsClients {
		client := wsClients[index]
		err := client.WriteMessage(ws.TextMessage, dataBytes)
		if err != nil {
			zap.L().Error("error on write to a client", zap.Error(err), zap.Any("client", client.LocalAddr().String()))
			client.Close()
			unregisterClient(client)
		}
	}
}

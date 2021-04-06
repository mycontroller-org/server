package mcwebsocket

import (
	"strings"

	ws "github.com/gorilla/websocket"
	"github.com/mycontroller-org/backend/v2/pkg/json"
	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	fieldML "github.com/mycontroller-org/backend/v2/pkg/model/field"
	gatewayML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	schedulerML "github.com/mycontroller-org/backend/v2/pkg/model/scheduler"
	sourceML "github.com/mycontroller-org/backend/v2/pkg/model/source"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	wsML "github.com/mycontroller-org/backend/v2/pkg/model/websocket"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	quickid "github.com/mycontroller-org/backend/v2/pkg/utils/quick_id"
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

func onEventReceive(event *busML.BusData) {
	status := eventsQueue.Produce(event)
	if !status {
		zap.L().Error("failed to post selected tasks on processor queue")
	}
}

func processEvent(item interface{}) {
	event := item.(*busML.BusData)
	topic := event.Topic

	var resourceData interface{}
	resourceType := ""

	switch {
	case strings.HasSuffix(topic, mcbus.TopicEventGateway):
		resourceData = &gatewayML.Config{}
		resourceType = "gateway"

	case strings.HasSuffix(topic, mcbus.TopicEventNode):
		resourceData = &nodeML.Node{}
		resourceType = "node"

	case strings.HasSuffix(topic, mcbus.TopicEventSource):
		resourceData = &sourceML.Source{}
		resourceType = "source"

	case strings.HasSuffix(topic, mcbus.TopicEventFieldSet):
		resourceData = &fieldML.Field{}
		resourceType = "field"

	case strings.HasSuffix(topic, mcbus.TopicEventTask):
		resourceData = &taskML.Config{}
		resourceType = "task"

	case strings.HasSuffix(topic, mcbus.TopicEventSchedule):
		resourceData = &schedulerML.Config{}
		resourceType = "schedule"

	case strings.HasSuffix(topic, mcbus.TopicEventHandler):
		resourceData = &handlerML.Config{}
		resourceType = "handler"

	default:
		return
	}

	err := event.ToStruct(resourceData)
	if err != nil {
		zap.L().Warn("Failed to convet to target type", zap.Error(err))
		return
	}

	resource := wsML.Resource{
		Type:     resourceType,
		Resource: resourceData,
		ID:       "", // TODO: add id of the resource
	}

	qID, err := quickid.GetQuickID(resource.Resource)
	if err != nil {
		zap.L().Error("error on getting quick id", zap.Error(err))
		return
	}

	resource.QuickID = qID

	zap.L().Debug("resource received", zap.Any("resource", resourceData))

	response := wsML.Response{
		Type: wsML.ResponseTypeResource,
		Data: resource,
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

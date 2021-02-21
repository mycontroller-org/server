package task

import (
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/model/event"
	"github.com/mycontroller-org/backend/v2/pkg/model/field"
	"github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	"github.com/mycontroller-org/backend/v2/pkg/model/node"
	"github.com/mycontroller-org/backend/v2/pkg/model/sensor"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

type resourceWrapper struct {
	ResourceType string
	Resource     interface{}
	Tasks        []taskML.Config
}

// event types
const (
	eventTypeGateway            = "gateway"
	eventTypeNode               = "node"
	eventTypeSensor             = "sensor"
	eventTypeSensorFieldSet     = "sensor_field/set"
	eventTypeSensorFieldRequest = "sensor_field/request"
)

const (
	eventListenerPreQueueLimit   = 1000
	eventListenerPostQueueLimit  = 1000
	eventListenerPreWorkerLimit  = 5
	eventListenerPostWorkerLimit = 1
	eventListenerPreQueueName    = "task_event_listener_pre"
	eventListenerPostQueueName   = "task_event_listener_post"
)

var (
	preEventsQueue          *queueUtils.Queue
	postEventsQueue         *queueUtils.Queue
	preEventsSubscriptionID = int64(0)
	preEventsTopic          = ""
)

// initEventListener events listener
func initEventListener() error {
	preEventsQueue = queueUtils.New(eventListenerPreQueueName, eventListenerPreQueueLimit, processPreEvent, eventListenerPreWorkerLimit)
	postEventsQueue = queueUtils.New(eventListenerPostQueueName, eventListenerPostQueueLimit, resourcePostProcessor, eventListenerPostWorkerLimit)

	// on message receive add it in to our local queue
	// TODO: update to listen all the events
	preEventsTopic = mcbus.FormatTopic(mcbus.TopicEventSensorFieldSet)
	sID, err := mcbus.Subscribe(preEventsTopic, onEventReceive)
	if err != nil {
		return err
	}
	preEventsSubscriptionID = sID
	return nil
}

func closeEventListener() error {
	err := mcbus.Unsubscribe(preEventsTopic, preEventsSubscriptionID)
	if err != nil {
		return err
	}
	preEventsQueue.Close()
	postEventsQueue.Close()
	return nil
}

func onEventReceive(event *event.Event) {
	status := preEventsQueue.Produce(event)
	if !status {
		zap.L().Warn("Failed to store the event into queue", zap.Any("event", event))
	}
}

func processPreEvent(item interface{}) {
	event := item.(*event.Event)
	topic := event.Topic

	var resource interface{}
	resourceType := ""

	switch {
	case strings.HasSuffix(topic, eventTypeGateway):
		resource = &gateway.Config{}
		resourceType = eventTypeGateway

	case strings.HasSuffix(topic, eventTypeNode):
		resource = &node.Node{}
		resourceType = eventTypeNode

	case strings.HasSuffix(topic, eventTypeSensor):
		resource = &sensor.Sensor{}
		resourceType = eventTypeSensor

	case strings.HasSuffix(topic, eventTypeSensorFieldSet):
		resource = &field.Field{}
		resourceType = eventTypeSensorFieldSet

	case strings.HasSuffix(topic, eventTypeSensorFieldRequest):
		resource = &field.Field{}
		resourceType = eventTypeSensorFieldRequest

	default:
		zap.L().Warn("unknown event", zap.Any("event", event))
		return
	}

	err := event.ToStruct(resource)
	if err != nil {
		zap.L().Warn("Failed to convet to target type", zap.Error(err))
		return
	}

	resourceWrapper := &resourceWrapper{ResourceType: resourceType, Resource: resource}
	err = resourcePreProcessor(resourceWrapper)
	if err != nil {
		zap.L().Error("Error on executing a resource", zap.Any("resource", resourceWrapper), zap.Error(err))
		return
	}

	if len(resourceWrapper.Tasks) > 0 {
		status := postEventsQueue.Produce(resourceWrapper)
		if !status {
			zap.L().Error("failed to post selected tasks on post processor queue")
		}
	}

}

func resourcePreProcessor(resource *resourceWrapper) error {
	zap.L().Debug("resource received", zap.Any("resource", resource))

	tasks := tasksStore.filterTasks(resource)
	zap.L().Debug("filtered", zap.Any("numberOftasks", len(tasks)))

	for index := 0; index < len(tasks); index++ {
		task := tasks[index]
		zap.L().Debug("executing task", zap.String("id", task.ID), zap.String("description", task.Description))
		if len(tasks) > 0 {
			resource.Tasks = tasks
		}
	}
	return nil
}

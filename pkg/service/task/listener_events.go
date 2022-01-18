package task

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
	dataRepositoryTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"github.com/mycontroller-org/server/v2/pkg/types/source"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

type eventWrapper struct {
	Event *eventTY.Event
	Tasks []taskTY.Config
}

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
	preEventsTopic          = "" // updated dynamically
)

// initEventListener events listener
func initEventListener() error {
	preEventsQueue = queueUtils.New(eventListenerPreQueueName, eventListenerPreQueueLimit, processPreEvent, eventListenerPreWorkerLimit)
	postEventsQueue = queueUtils.New(eventListenerPostQueueName, eventListenerPostQueueLimit, resourcePostProcessor, eventListenerPostWorkerLimit)

	// on message receive add it in to our local queue
	preEventsTopic = mcbus.FormatTopic(mcbus.TopicEventsAll)
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

func onEventReceive(busData *busTY.BusData) {
	status := preEventsQueue.Produce(busData)
	if !status {
		zap.L().Warn("failed to store the event into queue", zap.Any("event", busData))
	}
}

func processPreEvent(item interface{}) {
	busData := item.(*busTY.BusData)

	event := &eventTY.Event{}
	err := busData.LoadData(event)
	if err != nil {
		zap.L().Warn("error on convet to target type", zap.Any("topic", busData.Topic), zap.Error(err))
		return
	}

	var out interface{}

	// supported entity events
	switch event.EntityType {
	case types.EntityGateway:
		out = &gatewayTY.Config{}

	case types.EntityNode:
		out = &nodeTY.Node{}

	case types.EntitySource:
		out = &source.Source{}

	case types.EntityField:
		out = &fieldTY.Field{}

	case types.EntityDataRepository:
		out = &dataRepositoryTY.Config{}

	default:
		// return do not proceed further
		return
	}

	err = event.LoadEntity(out)
	if err != nil {
		zap.L().Warn("error on loading entity", zap.Any("event", event), zap.Error(err))
		return
	}
	event.Entity = out

	resourceWrapper := &eventWrapper{Event: event}
	err = resourcePreProcessor(resourceWrapper)
	if err != nil {
		zap.L().Error("error on executing a resource", zap.Any("resource", resourceWrapper), zap.Error(err))
		return
	}

	if len(resourceWrapper.Tasks) > 0 {
		status := postEventsQueue.Produce(resourceWrapper)
		if !status {
			zap.L().Error("failed to post selected tasks on post processor queue")
		}
	}
}

func resourcePreProcessor(evntWrapper *eventWrapper) error {
	zap.L().Debug("eventWrapper received", zap.Any("eventWrapper", evntWrapper))

	tasks := tasksStore.filterTasks(evntWrapper)
	zap.L().Debug("filtered", zap.Any("numberOftasks", len(tasks)))

	for index := 0; index < len(tasks); index++ {
		task := tasks[index]
		zap.L().Debug("executing task", zap.String("id", task.ID), zap.String("description", task.Description))
		if len(tasks) > 0 {
			evntWrapper.Tasks = tasks
		}
	}
	return nil
}

func resourcePostProcessor(item interface{}) {
	evntWrapper, ok := item.(*eventWrapper)
	if !ok {
		zap.L().Warn("supplied item is not resourceWrapper", zap.Any("item", item))
		return
	}

	zap.L().Debug("resourceWrapper received", zap.String("entityType", evntWrapper.Event.EntityType))

	for index := 0; index < len(evntWrapper.Tasks); index++ {
		task := evntWrapper.Tasks[index]
		executeTask(&task, evntWrapper)
	}
}

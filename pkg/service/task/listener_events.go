package task

import (
	types "github.com/mycontroller-org/server/v2/pkg/types"
	dataRepositoryTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"github.com/mycontroller-org/server/v2/pkg/types/source"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

type eventWrapper struct {
	Event *eventTY.Event
	Tasks []taskTY.Config
}

// initEventListener events listener
func (svc *TaskService) initEventListener() error {

	// on message receive add it in to our local queue
	sID, err := svc.bus.Subscribe(svc.preEventsQueue.Topic, svc.onEventReceive)
	if err != nil {
		return err
	}
	svc.preEventsQueue.SubscriptionId = sID
	return nil
}

func (svc *TaskService) closeEventListener() error {
	err := svc.bus.Unsubscribe(svc.preEventsQueue.Topic, svc.preEventsQueue.SubscriptionId)
	if err != nil {
		return err
	}
	svc.preEventsQueue.Close()
	svc.postEventsQueue.Close()
	return nil
}

func (svc *TaskService) onEventReceive(busData *busTY.BusData) {
	status := svc.preEventsQueue.Produce(busData)
	if !status {
		svc.logger.Warn("failed to store the event into queue", zap.Any("event", busData))
	}
}

func (svc *TaskService) processPreEvent(item interface{}) error {
	busData := item.(*busTY.BusData)

	event := &eventTY.Event{}
	err := busData.LoadData(event)
	if err != nil {
		svc.logger.Warn("error on convert to target type", zap.Any("topic", busData.Topic), zap.Error(err))
		return nil
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
		return nil
	}

	err = event.LoadEntity(out)
	if err != nil {
		svc.logger.Warn("error on loading entity", zap.Any("event", event), zap.Error(err))
		return nil
	}
	event.Entity = out

	resourceWrapper := &eventWrapper{Event: event}
	err = svc.resourcePreProcessor(resourceWrapper)
	if err != nil {
		svc.logger.Error("error on executing a resource", zap.Any("resource", resourceWrapper), zap.Error(err))
		return err
	}

	if len(resourceWrapper.Tasks) > 0 {
		status := svc.postEventsQueue.Produce(resourceWrapper)
		if !status {
			svc.logger.Error("failed to post selected tasks on post processor queue")
		}
	}
	return nil
}

func (svc *TaskService) resourcePreProcessor(evntWrapper *eventWrapper) error {
	svc.logger.Debug("eventWrapper received", zap.Any("eventWrapper", evntWrapper))

	tasks := svc.store.filterTasks(evntWrapper)
	svc.logger.Debug("filtered", zap.Any("numberOftasks", len(tasks)))

	for index := 0; index < len(tasks); index++ {
		task := tasks[index]
		svc.logger.Debug("executing task", zap.String("id", task.ID), zap.String("description", task.Description))
		if len(tasks) > 0 {
			evntWrapper.Tasks = tasks
		}
	}
	return nil
}

func (svc *TaskService) resourcePostProcessor(item interface{}) error {
	evntWrapper, ok := item.(*eventWrapper)
	if !ok {
		svc.logger.Warn("supplied item is not resourceWrapper", zap.Any("item", item))
		return nil
	}

	svc.logger.Debug("resourceWrapper received", zap.String("entityType", evntWrapper.Event.EntityType))

	for index := 0; index < len(evntWrapper.Tasks); index++ {
		task := evntWrapper.Tasks[index]
		svc.executeTask(&task, evntWrapper)
	}
	return nil
}

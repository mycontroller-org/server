package resource

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	"github.com/mycontroller-org/backend/v2/pkg/model/event"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

var (
	eventQueue   *queueUtils.Queue
	queueSize    = int(1000)
	queueWorkers = 5
)

// Init starts resource server listener
func Init() error {
	eventQueue = queueUtils.New("resource_service", queueSize, processEvent, queueWorkers)

	// on event receive add it in to our local queue
	_, err := mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicServiceResourceServer), onEvent)
	if err != nil {
		return err
	}

	return nil
}

// Close the service
func Close() {
	eventQueue.Close()
}

func onEvent(event *event.Event) {
	reqEvent := &rsModel.Event{}
	err := event.ToStruct(reqEvent)
	if err != nil {
		zap.L().Warn("Failed to convet to target type", zap.Error(err))
		return
	}

	if reqEvent == nil {
		zap.L().Warn("Received a nil event", zap.Any("event", event))
		return
	}
	zap.L().Debug("Event added into processing queue", zap.Any("event", reqEvent))
	status := eventQueue.Produce(reqEvent)
	if !status {
		zap.L().Warn("Failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processEvent from the queue
func processEvent(item interface{}) {
	request := item.(*rsModel.Event)
	zap.L().Debug("Processing an event", zap.Any("event", request))
	start := time.Now()
	switch request.Type {
	case rsModel.TypeGateway:
		err := gatewayService(request)
		if err != nil {
			zap.L().Error("error on serving gateway request", zap.Error(err))
		}

	case rsModel.TypeTask:
		err := taskService(request)
		if err != nil {
			zap.L().Error("error on serving task request", zap.Error(err))
		}

	case rsModel.TypeNotifyHandler:
		err := handlerService(request)
		if err != nil {
			zap.L().Error("error on serving handler request", zap.Error(err))
		}

	case rsModel.TypeScheduler:
		err := schedulerService(request)
		if err != nil {
			zap.L().Error("error on serving scheduler request", zap.Error(err))
		}

	case rsModel.TypeResourceActionBySelector:
		err := resourceActionService(request)
		if err != nil {
			zap.L().Error("error on serving resource quickID request", zap.Error(err))
		}

	default:
		zap.L().Warn("unknown event type", zap.Any("event", request))
	}
	zap.L().Info("completed a resource service", zap.String("timeTaken", time.Since(start).String()), zap.Any("data", request))
}

func postResponse(topic string, response *rsModel.Event) error {
	if topic == "" {
		return nil
	}
	return mcbus.Publish(topic, response)
}

func getLabelsFilter(labels cmap.CustomStringMap) []stgml.Filter {
	filters := make([]stgml.Filter, 0)
	for key, value := range labels {
		filter := stgml.Filter{Key: fmt.Sprintf("labels.%s", key), Operator: stgml.OperatorEqual, Value: value}
		filters = append(filters, filter)
	}
	return filters
}

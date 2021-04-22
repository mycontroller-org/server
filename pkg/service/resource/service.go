package resource

import (
	"fmt"
	"time"

	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

var (
	eventQueue     *queueUtils.Queue
	queueSize      = int(1000)
	queueWorkers   = 5
	subscriptionID = int64(0)
)

// Init starts resource server listener
func Init() error {
	eventQueue = queueUtils.New("resource_service", queueSize, processEvent, queueWorkers)

	// on event receive add it in to our local queue
	sID, err := mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicServiceResourceServer), onEvent)
	if err != nil {
		return err
	}
	subscriptionID = sID
	return nil
}

// Close the service
func Close() {
	err := mcbus.Unsubscribe(mcbus.FormatTopic(mcbus.TopicServiceResourceServer), subscriptionID)
	if err != nil {
		zap.L().Error("error on unsubscription", zap.Error(err))
	}
	eventQueue.Close()
}

func onEvent(data *busML.BusData) {
	reqEvent := &rsModel.ServiceEvent{}
	err := data.ToStruct(reqEvent)
	if err != nil {
		zap.L().Warn("Failed to convet to target type", zap.Error(err))
		return
	}

	if reqEvent == nil {
		zap.L().Warn("Received a nil event", zap.Any("event", data))
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
	request := item.(*rsModel.ServiceEvent)
	zap.L().Debug("Processing an event", zap.Any("event", request))
	start := time.Now()
	switch request.Type {
	case rsModel.TypeGateway:
		err := gatewayService(request)
		if err != nil {
			zap.L().Error("error on serving gateway request", zap.Error(err))
		}

	case rsModel.TypeNode:
		err := nodeService(request)
		if err != nil {
			zap.L().Error("error on serving node request", zap.Error(err))
		}

	case rsModel.TypeTask:
		err := taskService(request)
		if err != nil {
			zap.L().Error("error on serving task request", zap.Error(err))
		}

	case rsModel.TypeHandler:
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

	case rsModel.TypeFirmware:
		err := firmwareService(request)
		if err != nil {
			zap.L().Error("error on serving firmware service request", zap.Error(err))
		}

	default:
		zap.L().Warn("unknown event type", zap.Any("event", request))
	}
	zap.L().Debug("completed a resource service", zap.String("timeTaken", time.Since(start).String()), zap.Any("data", request))
}

func postResponse(topic string, response *rsModel.ServiceEvent) error {
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

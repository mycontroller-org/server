package resource

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

var (
	eventQueue     *queueUtils.Queue
	queueSize      = int(1000)
	queueWorkers   = 5
	subscriptionID = int64(0)
)

// Start starts resource server listener
func Start() error {
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

func onEvent(data *busTY.BusData) {
	reqEvent := &rsTY.ServiceEvent{}
	err := data.LoadData(reqEvent)
	if err != nil {
		zap.L().Warn("failed to covert to target type", zap.Error(err))
		return
	}

	if reqEvent.Type == "" {
		zap.L().Warn("received an empty event", zap.Any("event", data))
		return
	}
	zap.L().Debug("event added into processing queue", zap.Any("event", reqEvent))
	status := eventQueue.Produce(reqEvent)
	if !status {
		zap.L().Warn("failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processEvent from the queue
func processEvent(item interface{}) {
	request := item.(*rsTY.ServiceEvent)
	zap.L().Debug("processing an event", zap.Any("event", request))
	start := time.Now()
	switch request.Type {
	case rsTY.TypeGateway:
		err := gatewayService(request)
		if err != nil {
			zap.L().Error("error on serving gateway request", zap.Error(err))
		}

	case rsTY.TypeNode:
		err := nodeService(request)
		if err != nil {
			zap.L().Error("error on serving node request", zap.Error(err))
		}

	case rsTY.TypeTask:
		err := taskService(request)
		if err != nil {
			zap.L().Error("error on serving task request", zap.Error(err))
		}

	case rsTY.TypeHandler:
		err := handlerService(request)
		if err != nil {
			zap.L().Error("error on serving handler request", zap.Error(err))
		}

	case rsTY.TypeScheduler:
		err := schedulerService(request)
		if err != nil {
			zap.L().Error("error on serving scheduler request", zap.Error(err))
		}

	case rsTY.TypeResourceAction:
		err := resourceActionService(request)
		if err != nil {
			zap.L().Error("error on serving resource quickID request", zap.Error(err))
		}

	case rsTY.TypeFirmware:
		err := firmwareService(request)
		if err != nil {
			zap.L().Error("error on serving firmware service request", zap.Error(err))
		}

	default:
		zap.L().Warn("unknown event type", zap.Any("event", request))
	}
	zap.L().Debug("completed a resource service", zap.String("timeTaken", time.Since(start).String()), zap.Any("data", request))
}

func postResponse(topic string, response *rsTY.ServiceEvent) error {
	if topic == "" {
		return nil
	}
	return mcbus.Publish(topic, response)
}

func getLabelsFilter(labels cmap.CustomStringMap) []storageTY.Filter {
	filters := make([]storageTY.Filter, 0)
	for key, value := range labels {
		filter := storageTY.Filter{Key: fmt.Sprintf("labels.%s", key), Operator: storageTY.OperatorEqual, Value: value}
		filters = append(filters, filter)
	}
	return filters
}

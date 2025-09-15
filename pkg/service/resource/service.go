package resource

import (
	"context"
	"fmt"
	"time"

	actionAPI "github.com/mycontroller-org/server/v2/pkg/api/action"
	entityAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"

	"go.uber.org/zap"
)

const (
	defaultQueueSize = int(50)
	defaultWorkers   = int(5)
)

type ResourceService struct {
	ctx         context.Context
	logger      *zap.Logger
	api         *entityAPI.API
	actionAPI   *actionAPI.ActionAPI
	bus         busTY.Plugin
	eventsQueue *queueUtils.QueueSpec
}

func New(ctx context.Context) (serviceTY.Service, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	api, err := entityAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	_actionAPI, err := actionAPI.New(ctx)
	if err != nil {
		return nil, err
	}

	svc := &ResourceService{
		ctx:       ctx,
		logger:    logger.Named("resource_service"),
		api:       api,
		actionAPI: _actionAPI,
		bus:       bus,
	}

	svc.eventsQueue = &queueUtils.QueueSpec{
		Queue:          queueUtils.New(svc.logger, "resource_service", defaultQueueSize, svc.processEvent, defaultWorkers),
		Topic:          topic.TopicServiceResourceServer,
		SubscriptionId: -1,
	}

	return svc, nil
}

func (svc *ResourceService) Name() string {
	return "resource_service"
}

// Start starts resource server listener
func (svc *ResourceService) Start() error {
	// on event receive add it in to our local queue
	sID, err := svc.bus.Subscribe(svc.eventsQueue.Topic, svc.onEvent)
	if err != nil {
		return err
	}
	svc.eventsQueue.SubscriptionId = sID
	return nil
}

// Close the service
func (svc *ResourceService) Close() error {
	err := svc.bus.Unsubscribe(svc.eventsQueue.Topic, svc.eventsQueue.SubscriptionId)
	if err != nil {
		svc.logger.Error("error on unsubscription", zap.Error(err))
	}
	svc.eventsQueue.Close()
	return nil
}

func (svc *ResourceService) onEvent(data *busTY.BusData) {
	reqEvent := &rsTY.ServiceEvent{}
	err := data.LoadData(reqEvent)
	if err != nil {
		svc.logger.Warn("failed to covert to target type", zap.Error(err))
		return
	}

	if reqEvent.Type == "" {
		svc.logger.Warn("received an empty event", zap.Any("event", data))
		return
	}
	svc.logger.Debug("event added into processing queue", zap.Any("event", reqEvent))
	status := svc.eventsQueue.Produce(reqEvent)
	if !status {
		svc.logger.Warn("failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processEvent from the queue
func (svc *ResourceService) processEvent(item interface{}) error {
	request := item.(*rsTY.ServiceEvent)
	svc.logger.Debug("processing an event", zap.Any("event", request))
	start := time.Now()
	switch request.Type {
	case rsTY.TypeGateway:
		err := svc.gatewayService(request)
		if err != nil {
			svc.logger.Error("error on serving gateway request", zap.Error(err))
		}

	case rsTY.TypeNode:
		err := svc.nodeService(request)
		if err != nil {
			svc.logger.Error("error on serving node request", zap.Error(err))
		}

	case rsTY.TypeTask:
		err := svc.taskService(request)
		if err != nil {
			svc.logger.Error("error on serving task request", zap.Error(err))
		}

	case rsTY.TypeHandler:
		err := svc.handlerService(request)
		if err != nil {
			svc.logger.Error("error on serving handler request", zap.Error(err))
		}

	case rsTY.TypeScheduler:
		err := svc.schedulerService(request)
		if err != nil {
			svc.logger.Error("error on serving scheduler request", zap.Error(err))
		}

	case rsTY.TypeResourceAction:
		err := svc.resourceActionService(request)
		if err != nil {
			svc.logger.Error("error on serving resource quickID request", zap.Error(err))
		}

	case rsTY.TypeFirmware:
		err := svc.firmwareService(request)
		if err != nil {
			svc.logger.Error("error on serving firmware service request", zap.Error(err))
		}

	case rsTY.TypeVirtualAssistant:
		err := svc.virtualAssistantService(request)
		if err != nil {
			svc.logger.Error("error on serving virtual assistant service request", zap.Error(err))
		}

	default:
		svc.logger.Warn("unknown event type", zap.Any("event", request))
	}
	svc.logger.Debug("completed a resource service", zap.String("timeTaken", time.Since(start).String()), zap.Any("data", request))
	return nil
}

func (svc *ResourceService) postResponse(topic string, response *rsTY.ServiceEvent) error {
	if topic == "" {
		return nil
	}
	return svc.bus.Publish(topic, response)
}

func (svc *ResourceService) getLabelsFilter(labels cmap.CustomStringMap) []storageTY.Filter {
	filters := make([]storageTY.Filter, 0)
	for key, value := range labels {
		filter := storageTY.Filter{Key: fmt.Sprintf("labels.%s", key), Operator: storageTY.OperatorEqual, Value: value}
		filters = append(filters, filter)
	}
	return filters
}

package handler

import (
	"context"

	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

const (
	defaultQueueSize = int(100)
	defaultWorkers   = int(1)
)

type HandlerService struct {
	ctx          context.Context
	logger       *zap.Logger
	store        *Store
	filter       *sfTY.ServiceFilter
	bus          busTY.Plugin
	enc          *encryptionAPI.Encryption
	serviceQueue *queueUtils.QueueSpec
	messageQueue *queueUtils.QueueSpec
}

func New(ctx context.Context, filter *sfTY.ServiceFilter) (serviceTY.Service, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	enc, err := encryptionAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	svc := &HandlerService{
		ctx:    ctx,
		logger: logger.Named("handler_service"),
		filter: filter,
		bus:    bus,
		enc:    enc,
	}

	svc.store = &Store{handlers: make(map[string]handlerTY.Plugin), logger: svc.logger}

	svc.serviceQueue = &queueUtils.QueueSpec{
		Queue:          queueUtils.New(svc.logger, "handler_service", defaultQueueSize, svc.postProcessServiceEvent, defaultWorkers),
		Topic:          topic.TopicServiceHandler,
		SubscriptionId: -1,
	}

	svc.messageQueue = &queueUtils.QueueSpec{
		Queue:          queueUtils.New(svc.logger, "handler_message", defaultQueueSize, svc.processHandlerMessage, defaultWorkers),
		Topic:          topic.TopicPostMessageNotifyHandler,
		SubscriptionId: -1,
	}

	return svc, nil
}

func (svc *HandlerService) Name() string {
	return "handler_service"
}

// Start handler service listener
func (svc *HandlerService) Start() error {
	if svc.filter.Disabled {
		svc.logger.Info("handler service disabled")
		return nil
	}

	if svc.filter.HasFilter() {
		svc.logger.Info("handler service filter config", zap.Any("filter", svc.filter))
	} else {
		svc.logger.Debug("there is no filter applied to handler service")
	}

	// on message receive add it in to our local queue
	id, err := svc.bus.Subscribe(svc.serviceQueue.Topic, svc.onServiceEvent)
	if err != nil {
		return err
	}
	svc.serviceQueue.SubscriptionId = id

	err = svc.initMessageListener()
	if err != nil {
		return err
	}

	// load handlers
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeHandler,
		Command: rsTY.CommandLoadAll,
	}
	return svc.bus.Publish(topic.TopicServiceResourceServer, reqEvent)
}

// Close the service listener
func (svc *HandlerService) Close() error {
	if svc.filter.Disabled {
		return nil
	}
	svc.unloadAll()
	svc.serviceQueue.Close()
	svc.closeMessageListener()
	return nil
}

func (svc *HandlerService) onServiceEvent(event *busTY.BusData) {
	reqEvent := &rsTY.ServiceEvent{}
	err := event.LoadData(reqEvent)
	if err != nil {
		svc.logger.Warn("failed to convert to target type", zap.Error(err))
		return
	}
	if reqEvent.Type == "" {
		svc.logger.Warn("received an empty event", zap.Any("event", event))
		return
	}
	svc.logger.Debug("event added into processing queue", zap.Any("event", reqEvent))
	status := svc.serviceQueue.Produce(reqEvent)
	if !status {
		svc.logger.Warn("failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// postProcessServiceEvent from the queue
func (svc *HandlerService) postProcessServiceEvent(event interface{}) {
	reqEvent := event.(*rsTY.ServiceEvent)
	svc.logger.Debug("processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsTY.TypeHandler {
		svc.logger.Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsTY.CommandStart:
		cfg := svc.getConfig(reqEvent)
		if cfg != nil && helper.IsMine(svc.filter, cfg.Type, cfg.ID, cfg.Labels) {
			err := svc.startHandler(cfg)
			if err != nil {
				svc.logger.Error("error on starting a handler", zap.Error(err), zap.String("handler", cfg.ID))
			}
		}

	case rsTY.CommandStop:
		if reqEvent.ID != "" {
			err := svc.stopHandler(reqEvent.ID)
			if err != nil {
				svc.logger.Error("error on stopping a service", zap.Error(err))
			}
			return
		}
		cfg := svc.getConfig(reqEvent)
		if cfg != nil {
			err := svc.stopHandler(cfg.ID)
			if err != nil {
				svc.logger.Error("error on stopping a service", zap.Error(err))
			}
		}

	case rsTY.CommandReload:
		cfg := svc.getConfig(reqEvent)
		if cfg != nil && helper.IsMine(svc.filter, cfg.Type, cfg.ID, cfg.Labels) {
			err := svc.reloadHandler(cfg)
			if err != nil {
				svc.logger.Error("error on reload a service", zap.Error(err))
			}
		}

	case rsTY.CommandUnloadAll:
		svc.unloadAll()

	default:
		svc.logger.Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func (svc *HandlerService) getConfig(reqEvent *rsTY.ServiceEvent) *handlerTY.Config {
	cfg := &handlerTY.Config{}
	err := reqEvent.LoadData(cfg)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil
	}
	return cfg
}

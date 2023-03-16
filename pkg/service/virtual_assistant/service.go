package service

import (
	"context"

	"github.com/gorilla/mux"
	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	vaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/types"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

const (
	defaultQueueSize = int(50)
	defaultWorkers   = int(1)
)

type VirtualAssistantService struct {
	ctx         context.Context
	logger      *zap.Logger
	filter      *sfTY.ServiceFilter
	store       *Store
	bus         busTY.Plugin
	enc         *encryptionAPI.Encryption
	eventsQueue *queueUtils.QueueSpec
	router      *mux.Router
}

func New(ctx context.Context, filter *sfTY.ServiceFilter, router *mux.Router) (serviceTY.Service, error) {
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
	if filter == nil {
		filter = &sfTY.ServiceFilter{}
	}

	svc := &VirtualAssistantService{
		ctx:    ctx,
		logger: logger.Named("virtual_assistant_service"),
		filter: filter,
		bus:    bus,
		enc:    enc,
		router: router,
	}

	svc.store = &Store{services: make(map[string]vaTY.Plugin)}

	svc.eventsQueue = &queueUtils.QueueSpec{
		Queue:          queueUtils.New(svc.logger, "virtual_assistant_service", defaultQueueSize, svc.processEvent, defaultWorkers),
		Topic:          topic.TopicServiceVirtualAssistant,
		SubscriptionId: -1,
	}

	// register handler path
	// needs to be registered before passing into http_router
	svc.registerServiceRoute()

	return svc, nil
}

func (svc *VirtualAssistantService) Name() string {
	return "virtual_assistant_service"
}

// Start starts resource server listener
func (svc *VirtualAssistantService) Start() error {
	if svc.filter.Disabled {
		svc.logger.Info("virtual assistant service disabled")
		return nil
	}

	if svc.filter.HasFilter() {
		svc.logger.Info("virtual assistant service filter config", zap.Any("filter", svc.filter))
	} else {
		svc.logger.Debug("there is no filter applied to virtual assistant service")
	}

	// on event receive add it in to our local queue
	sId, err := svc.bus.Subscribe(svc.eventsQueue.Topic, svc.onEvent)
	if err != nil {
		svc.logger.Error("error on subscription", zap.Error(err))
		return err
	}
	svc.eventsQueue.SubscriptionId = sId

	// load virtual assistants
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeVirtualAssistant,
		Command: rsTY.CommandLoadAll,
	}
	err = svc.bus.Publish(topic.TopicServiceResourceServer, reqEvent)
	if err != nil {
		return err
	}

	return nil
}

// Close the service
func (svc *VirtualAssistantService) Close() error {
	if svc.filter.Disabled {
		return nil
	}
	svc.unloadAll()
	svc.eventsQueue.Close()
	return nil
}

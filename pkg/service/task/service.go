package task

import (
	"context"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

const (
	defaultQueueSize           = int(100)
	defaultPreEventsQueueSize  = int(1000)
	defaultPostEventsQueueSize = int(1000)
	defaultWorkers             = int(1)
	defaultPreEventsWorkers    = int(5)
	defaultPostEventsWorkers   = int(1)

	defaultScriptTimeout = "30m" // default script timeout 30 minutes
)

type TaskService struct {
	ctx             context.Context
	logger          *zap.Logger
	bus             busTY.Plugin
	scheduler       schedulerTY.CoreScheduler
	filter          *sfTY.ServiceFilter
	serviceQueue    *queueUtils.QueueSpec
	preEventsQueue  *queueUtils.QueueSpec
	postEventsQueue *queueUtils.QueueSpec
	variablesEngine types.VariablesEngine
	store           *Store
}

func New(ctx context.Context, filter *sfTY.ServiceFilter, variablesEngine types.VariablesEngine) (serviceTY.Service, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	scheduler, err := schedulerTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	if filter == nil {
		filter = &sfTY.ServiceFilter{}
	}

	svc := &TaskService{
		ctx:             ctx,
		logger:          logger.Named("task_service"),
		bus:             bus,
		scheduler:       scheduler,
		variablesEngine: variablesEngine,
		filter:          filter,
	}

	svc.store = &Store{
		tasks:        make(map[string]taskTY.Config),
		pollingTasks: make([]string, 0),
		logger:       svc.logger,
		bus:          svc.bus,
	}

	svc.serviceQueue = &queueUtils.QueueSpec{
		Queue:          queueUtils.New(svc.logger, "task_service", defaultQueueSize, svc.processServiceEvent, defaultWorkers),
		Topic:          topic.TopicServiceTask,
		SubscriptionId: -1,
	}

	svc.preEventsQueue = &queueUtils.QueueSpec{
		Queue:          queueUtils.New(svc.logger, "task_events_pre", defaultPreEventsQueueSize, svc.processPreEvent, defaultPreEventsWorkers),
		Topic:          topic.TopicEventsAll,
		SubscriptionId: -1,
	}

	svc.postEventsQueue = &queueUtils.QueueSpec{
		Queue:          queueUtils.New(svc.logger, "task_events_post", defaultPostEventsQueueSize, svc.resourcePostProcessor, defaultPostEventsWorkers),
		Topic:          "_no-topic-internal_",
		SubscriptionId: -1,
	}

	return svc, nil
}

func (svc *TaskService) Name() string {
	return "task_service"
}

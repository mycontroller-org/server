package systemjobs

import (
	"context"

	entityAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

const (
	defaultQueueSize = int(50)
	defaultWorkers   = int(1)
)

type SystemJobsService struct {
	ctx         context.Context
	logger      *zap.Logger
	scheduler   schedulerTY.CoreScheduler
	api         *entityAPI.API
	bus         busTY.Plugin
	eventsQueue *queueUtils.QueueSpec
}

func New(ctx context.Context) (serviceTY.Service, error) {
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
	api, err := entityAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	svc := &SystemJobsService{
		ctx:       ctx,
		logger:    logger.Named("system_jobs_service"),
		scheduler: scheduler,
		api:       api,
		bus:       bus,
	}

	svc.eventsQueue = &queueUtils.QueueSpec{
		Queue:          queueUtils.New(svc.logger, "system_jobs_service", defaultQueueSize, svc.processEvent, defaultWorkers),
		Topic:          topic.TopicInternalSystemJobs,
		SubscriptionId: -1,
	}

	return svc, nil
}

func (svc *SystemJobsService) Name() string {
	return "system_jobs_service"
}

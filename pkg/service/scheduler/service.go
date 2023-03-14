package scheduler

import (
	"context"
	"fmt"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

const (
	schedulePrefix       = "user_schedule"
	defaultQueueSize     = int(100)
	defaultWorkers       = int(1)
	defaultScriptTimeout = "30m" // default script timeout 30 minutes
)

type SchedulerService struct {
	ctx             context.Context
	logger          *zap.Logger
	coreScheduler   schedulerTY.CoreScheduler
	bus             busTY.Plugin
	sunriseApi      types.Sunrise
	filter          *sfTY.ServiceFilter
	eventsQueue     *queueUtils.QueueSpec
	variablesEngine types.VariablesEngine
}

func New(ctx context.Context, filter *sfTY.ServiceFilter, variablesEngine types.VariablesEngine, sunriseApi types.Sunrise) (serviceTY.Service, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	coreScheduler, err := schedulerTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	if filter == nil {
		filter = &sfTY.ServiceFilter{}
	}

	svc := &SchedulerService{
		ctx:             ctx,
		logger:          logger.Named("scheduler_service"),
		coreScheduler:   coreScheduler,
		bus:             bus,
		sunriseApi:      sunriseApi,
		variablesEngine: variablesEngine,
		filter:          filter,
	}

	svc.eventsQueue = &queueUtils.QueueSpec{
		Queue:          queueUtils.New(svc.logger, "scheduler_service", defaultQueueSize, svc.processServiceEvent, defaultWorkers),
		Topic:          topic.TopicServiceScheduler,
		SubscriptionId: -1,
	}

	return svc, nil
}

func (svc *SchedulerService) Name() string {
	return "scheduler_service"
}

func (svc *SchedulerService) schedule(cfg *schedulerTY.Config) {
	if cfg.State == nil {
		cfg.State = &schedulerTY.State{}
	}
	name := getScheduleID(cfg.ID)

	// update validity, if it is on date job, add date(with year) on validity
	err := updateOnDateJobValidity(cfg)
	if err != nil {
		svc.logger.Error("error on updating on date job config", zap.Error(err), zap.String("id", cfg.ID))
		return
	}

	cronSpec, err := svc.getCronSpec(cfg)
	if err != nil {
		svc.logger.Error("error on creating cron spec", zap.Error(err))
		cfg.State.LastStatus = false
		cfg.State.Message = fmt.Sprintf("Error on cron spec creation: %s", err.Error())
		busUtils.SetScheduleState(svc.logger, svc.bus, cfg.ID, *cfg.State)
		return
	}
	err = svc.coreScheduler.AddFunc(name, cronSpec, svc.getScheduleTriggerFunc(cfg, cronSpec))
	if err != nil {
		svc.logger.Error("error on adding schedule", zap.Error(err))
		cfg.State.LastStatus = false
		cfg.State.Message = fmt.Sprintf("Error on adding into scheduler: %s", err.Error())
		busUtils.SetScheduleState(svc.logger, svc.bus, cfg.ID, *cfg.State)
	}
	svc.logger.Debug("added a schedule", zap.String("name", name), zap.String("ID", cfg.ID), zap.Any("cronSpec", cronSpec))
	cfg.State.Message = fmt.Sprintf("Added into scheduler. cron spec:[%s]", cronSpec)
	busUtils.SetScheduleState(svc.logger, svc.bus, cfg.ID, *cfg.State)
}

func (svc *SchedulerService) unschedule(id string) {
	name := getScheduleID(id)
	svc.coreScheduler.RemoveFunc(name)
	svc.logger.Debug("removed a schedule", zap.String("name", name), zap.String("id", id))
}

func (svc *SchedulerService) unloadAll() {
	svc.coreScheduler.RemoveWithPrefix(schedulePrefix)
}

func getScheduleID(id string) string {
	return fmt.Sprintf("%s_%s", schedulePrefix, id)
}

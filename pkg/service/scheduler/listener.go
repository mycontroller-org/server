package scheduler

import (
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

// Start scheduler service listener
func (svc *SchedulerService) Start() error {
	if svc.filter.Disabled {
		svc.logger.Info("schedule service disabled")
		return nil
	}

	if svc.filter.HasFilter() {
		svc.logger.Info("schedule service filter config", zap.Any("filter", svc.filter))
	} else {
		svc.logger.Debug("there is no filter applied to schedule service")
	}

	// on message receive add it in to our local queue
	id, err := svc.bus.Subscribe(svc.eventsQueue.Topic, svc.onServiceEvent)
	if err != nil {
		return err
	}
	svc.eventsQueue.SubscriptionId = id

	svc.logger.Debug("scheduler started", zap.Any("filter", svc.filter))

	// load initial schedules, request via bus
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeScheduler,
		Command: rsTY.CommandLoadAll,
	}
	return svc.bus.Publish(topic.TopicServiceResourceServer, reqEvent)
}

// Close the service listener
func (svc *SchedulerService) Close() error {
	if svc.filter.Disabled {
		return nil
	}
	svc.unloadAll()
	svc.eventsQueue.Close()
	// close core scheduler
	svc.coreScheduler.Close()
	return nil
}

func (svc *SchedulerService) onServiceEvent(event *busTY.BusData) {
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
	status := svc.eventsQueue.Produce(reqEvent)
	if !status {
		svc.logger.Warn("failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processServiceEvent from the queue
func (svc *SchedulerService) processServiceEvent(event interface{}) error {
	reqEvent := event.(*rsTY.ServiceEvent)
	svc.logger.Debug("processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsTY.TypeScheduler {
		svc.logger.Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	// TODO: implement cancel of the script execution if any on the existing scheduler
	switch reqEvent.Command {
	case rsTY.CommandAdd:
		cfg := svc.getConfig(reqEvent)
		if cfg != nil && filterUtils.IsMine(svc.filter, cfg.Type, cfg.ID, cfg.Labels) {
			svc.schedule(cfg)
		}

	case rsTY.CommandRemove:
		cfg := svc.getConfig(reqEvent)
		if cfg != nil {
			svc.unschedule(cfg.ID)
		}

	case rsTY.CommandReload:
		cfg := svc.getConfig(reqEvent)
		if cfg != nil && filterUtils.IsMine(svc.filter, cfg.Type, cfg.ID, cfg.Labels) {
			svc.unschedule(cfg.ID)
			svc.schedule(cfg)
		}

	case rsTY.CommandUnloadAll:
		svc.unloadAll()

	default:
		svc.logger.Warn("unsupported command", zap.Any("event", reqEvent))
	}
	return nil
}

func (svc *SchedulerService) getConfig(reqEvent *rsTY.ServiceEvent) *schedulerTY.Config {
	cfg := &schedulerTY.Config{}
	err := reqEvent.LoadData(cfg)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil
	}
	return cfg
}

package task

import (
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

// Start task service listener
func (svc *TaskService) Start() error {
	if svc.filter.Disabled {
		svc.logger.Info("task service disabled")
		return nil
	}

	if svc.filter.HasFilter() {
		svc.logger.Info("task service filter config", zap.Any("filter", svc.filter))
	} else {
		svc.logger.Debug("there is no filter applied to task service")
	}

	// on message receive add it in to our local queue
	id, err := svc.bus.Subscribe(svc.serviceQueue.Topic, svc.onServiceEvent)
	if err != nil {
		return err
	}
	svc.serviceQueue.SubscriptionId = id

	err = svc.initEventListener()
	if err != nil {
		return err
	}

	// load tasks
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeTask,
		Command: rsTY.CommandLoadAll,
	}
	return svc.bus.Publish(topic.TopicServiceResourceServer, reqEvent)
}

// Close the service listener
func (svc *TaskService) Close() error {
	if svc.filter.Disabled {
		return nil
	}
	err := svc.closeEventListener()
	if err != nil {
		svc.logger.Error("error on closing event listener", zap.Error(err))
	}
	svc.store.RemoveAll()
	svc.unscheduleAll("") // remove all the scheduled tasks
	svc.serviceQueue.Close()
	return nil
}

func (svc *TaskService) onServiceEvent(busData *busTY.BusData) {
	reqEvent := &rsTY.ServiceEvent{}
	err := busData.LoadData(reqEvent)
	if err != nil {
		svc.logger.Warn("failed to convert to target type", zap.Error(err))
		return
	}
	if reqEvent.Type == "" {
		svc.logger.Warn("received an empty event", zap.Any("event", busData))
		return
	}
	svc.logger.Debug("event added into processing queue", zap.Any("event", reqEvent))
	status := svc.serviceQueue.Produce(reqEvent)
	if !status {
		svc.logger.Warn("failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processServiceEvent from the queue
func (svc *TaskService) processServiceEvent(event interface{}) {
	reqEvent := event.(*rsTY.ServiceEvent)
	svc.logger.Debug("processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsTY.TypeTask {
		svc.logger.Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	// TODO: implement cancel of the script execution if any on the existing task
	switch reqEvent.Command {
	case rsTY.CommandAdd:
		cfg := svc.getConfig(reqEvent)
		if cfg != nil {
			svc.store.Remove(cfg.ID, svc.unscheduleAll, svc.getScheduleId)
		}
		if cfg != nil && filterUtils.IsMine(svc.filter, cfg.EvaluationType, cfg.ID, cfg.Labels) {
			svc.store.Add(*cfg, svc.schedule)
		}

	case rsTY.CommandRemove:
		cfg := svc.getConfig(reqEvent)
		if cfg != nil {
			svc.store.Remove(cfg.ID, svc.unscheduleAll, svc.getScheduleId)
		}

	case rsTY.CommandUnloadAll:
		svc.store.RemoveAll()

	case rsTY.CommandReload:
		cfg := svc.getConfig(reqEvent)
		if cfg != nil {
			svc.store.Remove(cfg.ID, svc.unscheduleAll, svc.getScheduleId)
			if filterUtils.IsMine(svc.filter, cfg.EvaluationType, cfg.ID, cfg.Labels) {
				svc.store.Add(*cfg, svc.schedule)
			}
		}

	default:
		svc.logger.Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func (svc *TaskService) getConfig(reqEvent *rsTY.ServiceEvent) *taskTY.Config {
	cfg := &taskTY.Config{}
	err := reqEvent.LoadData(cfg)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil
	}

	return cfg
}

package task

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

const (
	serviceMessageQueueLimit = 100
	serviceMessageQueueName  = "service_listener_tasks"
)

var (
	tasksQueue *queueUtils.Queue
	svcFilter  *sfTY.ServiceFilter
)

// Start task service listener
func Start(filter *sfTY.ServiceFilter) error {
	svcFilter = filter
	if svcFilter.Disabled {
		zap.L().Info("task service disabled")
		return nil
	}

	if svcFilter.HasFilter() {
		zap.L().Info("task service filter config", zap.Any("filter", svcFilter))
	} else {
		zap.L().Debug("there is no filter applied to task service")
	}

	tasksQueue = queueUtils.New(serviceMessageQueueName, serviceMessageQueueLimit, processServiceEvent, 1)

	// on message receive add it in to our local queue
	_, err := mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicServiceTask), onServiceEvent)
	if err != nil {
		return err
	}

	err = initEventListener()
	if err != nil {
		return err
	}

	// load tasks
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeTask,
		Command: rsTY.CommandLoadAll,
	}
	topicResourceServer := mcbus.FormatTopic(mcbus.TopicServiceResourceServer)
	return mcbus.Publish(topicResourceServer, reqEvent)
}

// Close the service listener
func Close() {
	if svcFilter.Disabled {
		return
	}
	err := closeEventListener()
	if err != nil {
		zap.L().Error("error on closing event listener", zap.Error(err))
	}
	tasksStore.RemoveAll()
	tasksQueue.Close()
}

func onServiceEvent(busData *busTY.BusData) {
	reqEvent := &rsTY.ServiceEvent{}
	err := busData.LoadData(reqEvent)
	if err != nil {
		zap.L().Warn("failed to convet to target type", zap.Error(err))
		return
	}
	if reqEvent == nil {
		zap.L().Warn("received a nil message", zap.Any("event", busData))
		return
	}
	zap.L().Debug("event added into processing queue", zap.Any("event", reqEvent))
	status := tasksQueue.Produce(reqEvent)
	if !status {
		zap.L().Warn("failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processServiceEvent from the queue
func processServiceEvent(event interface{}) {
	reqEvent := event.(*rsTY.ServiceEvent)
	zap.L().Debug("processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsTY.TypeTask {
		zap.L().Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsTY.CommandAdd:
		cfg := getConfig(reqEvent)
		if cfg != nil {
			tasksStore.Remove(cfg.ID)
		}
		if cfg != nil && filterUtils.IsMine(svcFilter, cfg.EvaluationType, cfg.ID, cfg.Labels) {
			tasksStore.Add(*cfg)
		}

	case rsTY.CommandRemove:
		cfg := getConfig(reqEvent)
		if cfg != nil {
			tasksStore.Remove(cfg.ID)
		}

	case rsTY.CommandUnloadAll:
		tasksStore.RemoveAll()

	case rsTY.CommandReload:
		cfg := getConfig(reqEvent)
		if cfg != nil {
			tasksStore.Remove(cfg.ID)
			if filterUtils.IsMine(svcFilter, cfg.EvaluationType, cfg.ID, cfg.Labels) {
				tasksStore.Add(*cfg)
			}
		}

	default:
		zap.L().Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func getConfig(reqEvent *rsTY.ServiceEvent) *taskTY.Config {
	cfg := &taskTY.Config{}
	err := reqEvent.LoadData(cfg)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil
	}

	return cfg
}

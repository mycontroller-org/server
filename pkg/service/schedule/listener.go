package schedule

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	filterUtils "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

const (
	serviceMessageQueueLimit = 100
	serviceMessageQueueName  = "service_listener_scheduler"
)

var (
	serviceQueue *queueUtils.Queue
	svcFilter    *sfTY.ServiceFilter
)

// Start scheduler service listener
func Start(filter *sfTY.ServiceFilter) error {
	svcFilter = filter
	if svcFilter.Disabled {
		zap.L().Info("schedule service disabled")
		return nil
	}

	if svcFilter.HasFilter() {
		zap.L().Info("schedule service filter config", zap.Any("filter", svcFilter))
	} else {
		zap.L().Debug("there is no filter applied to schedule service")
	}

	serviceQueue = queueUtils.New(serviceMessageQueueName, serviceMessageQueueLimit, processServiceEvent, 1)

	// on message receive add it in to our local queue
	_, err := mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicServiceScheduler), onServiceEvent)
	if err != nil {
		return err
	}

	zap.L().Debug("scheduler started", zap.Any("filter", svcFilter))
	// load schedulers
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeScheduler,
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
	unloadAll()
	serviceQueue.Close()
}

func onServiceEvent(event *busTY.BusData) {
	reqEvent := &rsTY.ServiceEvent{}
	err := event.LoadData(reqEvent)
	if err != nil {
		zap.L().Warn("failed to convet to target type", zap.Error(err))
		return
	}
	if reqEvent == nil {
		zap.L().Warn("received a nil message", zap.Any("event", event))
		return
	}
	zap.L().Debug("event added into processing queue", zap.Any("event", reqEvent))
	status := serviceQueue.Produce(reqEvent)
	if !status {
		zap.L().Warn("failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processServiceEvent from the queue
func processServiceEvent(event interface{}) {
	reqEvent := event.(*rsTY.ServiceEvent)
	zap.L().Debug("processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsTY.TypeScheduler {
		zap.L().Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsTY.CommandAdd:
		cfg := getConfig(reqEvent)
		if cfg != nil && filterUtils.IsMine(svcFilter, cfg.Type, cfg.ID, cfg.Labels) {
			schedule(cfg)
		}

	case rsTY.CommandRemove:
		cfg := getConfig(reqEvent)
		if cfg != nil {
			unschedule(cfg.ID)
		}

	case rsTY.CommandReload:
		cfg := getConfig(reqEvent)
		if cfg != nil && filterUtils.IsMine(svcFilter, cfg.Type, cfg.ID, cfg.Labels) {
			unschedule(cfg.ID)
			schedule(cfg)
		}

	case rsTY.CommandUnloadAll:
		unloadAll()

	default:
		zap.L().Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func getConfig(reqEvent *rsTY.ServiceEvent) *scheduleTY.Config {
	cfg := &scheduleTY.Config{}
	err := reqEvent.LoadData(cfg)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil
	}
	return cfg
}

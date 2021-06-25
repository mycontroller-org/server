package schedule

import (
	busML "github.com/mycontroller-org/server/v2/pkg/model/bus"
	rsML "github.com/mycontroller-org/server/v2/pkg/model/resource_service"
	scheduleML "github.com/mycontroller-org/server/v2/pkg/model/schedule"
	sfML "github.com/mycontroller-org/server/v2/pkg/model/service_filter"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

const (
	serviceMessageQueueLimit = 100
	serviceMessageQueueName  = "service_listener_scheduler"
)

var (
	serviceQueue *queueUtils.Queue
	svcFilter    *sfML.ServiceFilter
)

// Start scheduler service listener
func Start(filter *sfML.ServiceFilter) error {
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
	reqEvent := rsML.ServiceEvent{
		Type:    rsML.TypeScheduler,
		Command: rsML.CommandLoadAll,
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

func onServiceEvent(event *busML.BusData) {
	reqEvent := &rsML.ServiceEvent{}
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
	reqEvent := event.(*rsML.ServiceEvent)
	zap.L().Debug("processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsML.TypeScheduler {
		zap.L().Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsML.CommandAdd:
		cfg := getConfig(reqEvent)
		if cfg != nil && helper.IsMine(svcFilter, cfg.Type, cfg.ID, cfg.Labels) {
			schedule(cfg)
		}

	case rsML.CommandRemove:
		cfg := getConfig(reqEvent)
		if cfg != nil {
			unschedule(cfg.ID)
		}

	case rsML.CommandReload:
		cfg := getConfig(reqEvent)
		if cfg != nil && helper.IsMine(svcFilter, cfg.Type, cfg.ID, cfg.Labels) {
			unschedule(cfg.ID)
			schedule(cfg)
		}

	case rsML.CommandUnloadAll:
		unloadAll()

	default:
		zap.L().Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func getConfig(reqEvent *rsML.ServiceEvent) *scheduleML.Config {
	cfg := &scheduleML.Config{}
	err := reqEvent.LoadData(cfg)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil
	}
	return cfg
}

package scheduler

import (
	"fmt"

	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	schedulerML "github.com/mycontroller-org/backend/v2/pkg/model/scheduler"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

const (
	serviceMessageQueueLimit = 100
	serviceMessageQueueName  = "service_listener_scheduler"
)

// Config of scheduler service
type Config struct {
	IDs    []string
	Labels cmap.CustomStringMap
}

var (
	serviceQueue *queueUtils.Queue
	svcCFG       *Config
)

// Init scheduler service listener
func Init(config cmap.CustomMap) error {
	svcCFG = &Config{}
	err := utils.MapToStruct(utils.TagNameNone, config, svcCFG)
	if err != nil {
		return err
	}

	serviceQueue = queueUtils.New(serviceMessageQueueName, serviceMessageQueueLimit, processServiceEvent, 1)

	// on message receive add it in to our local queue
	_, err = mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicServiceScheduler), onServiceEvent)
	if err != nil {
		return err
	}

	zap.L().Debug("Scheduler started", zap.Any("config", svcCFG))
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
	unloadAll()
	serviceQueue.Close()
}

func onServiceEvent(event *busML.BusData) {
	reqEvent := &rsML.ServiceEvent{}
	err := event.ToStruct(reqEvent)
	if err != nil {
		zap.L().Warn("Failed to convet to target type", zap.Error(err))
		return
	}
	if reqEvent == nil {
		zap.L().Warn("Received a nil message", zap.Any("event", event))
		return
	}
	zap.L().Debug("Event added into processing queue", zap.Any("event", reqEvent))
	status := serviceQueue.Produce(reqEvent)
	if !status {
		zap.L().Warn("Failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processServiceEvent from the queue
func processServiceEvent(event interface{}) {
	reqEvent := event.(*rsML.ServiceEvent)
	zap.L().Debug("Processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsML.TypeScheduler {
		zap.L().Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsML.CommandAdd:
		cfg := getConfig(reqEvent)
		if cfg != nil && helper.IsMine(svcCFG.IDs, svcCFG.Labels, cfg.ID, cfg.Labels) {
			schedule(cfg)
		}

	case rsML.CommandRemove:
		cfg := getConfig(reqEvent)
		if cfg != nil {
			unschedule(cfg.ID)
		}

	case rsML.CommandReload:
		cfg := getConfig(reqEvent)
		if cfg != nil && helper.IsMine(svcCFG.IDs, svcCFG.Labels, cfg.ID, cfg.Labels) {
			unschedule(cfg.ID)
			schedule(cfg)
		}

	case rsML.CommandUnloadAll:
		unloadAll()

	default:
		zap.L().Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func getConfig(reqEvent *rsML.ServiceEvent) *schedulerML.Config {
	cfg, ok := reqEvent.GetData().(schedulerML.Config)
	if !ok {
		zap.L().Error("error on data conversion", zap.String("receivedType", fmt.Sprintf("%T", reqEvent.GetData())))
		return nil
	}
	return &cfg
}

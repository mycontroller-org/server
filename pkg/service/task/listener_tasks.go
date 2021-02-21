package task

import (
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	"github.com/mycontroller-org/backend/v2/pkg/model/event"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

const (
	serviceMessageQueueLimit = 100
	serviceMessageQueueName  = "service_listener_tasks"
)

// Config of task service
type Config struct {
	IDs    []string
	Labels cmap.CustomStringMap
}

var (
	tasksQueue *queueUtils.Queue
	svcCFG     *Config
)

// Init task service listener
func Init(config cmap.CustomMap) error {
	svcCFG = &Config{}
	err := utils.MapToStruct(utils.TagNameNone, config, svcCFG)
	if err != nil {
		return err
	}

	tasksQueue = queueUtils.New(serviceMessageQueueName, serviceMessageQueueLimit, processServiceEvent, 1)

	// on message receive add it in to our local queue
	_, err = mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicServiceTask), onServiceEvent)
	if err != nil {
		return err
	}

	err = initEventListener()
	if err != nil {
		return err
	}

	// load tasks
	reqEvent := rsML.Event{
		Type:    rsML.TypeTask,
		Command: rsML.CommandLoadAll,
	}
	topicResourceServer := mcbus.FormatTopic(mcbus.TopicServiceResourceServer)
	return mcbus.Publish(topicResourceServer, reqEvent)
}

// Close the service listener
func Close() {
	err := closeEventListener()
	if err != nil {
		zap.L().Error("error on closing event listener", zap.Error(err))
	}
	tasksStore.RemoveAll()
	tasksQueue.Close()
}

func onServiceEvent(event *event.Event) {
	reqEvent := &rsML.Event{}
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
	status := tasksQueue.Produce(reqEvent)
	if !status {
		zap.L().Warn("Failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processServiceEvent from the queue
func processServiceEvent(event interface{}) {
	reqEvent := event.(*rsML.Event)
	zap.L().Debug("Processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsML.TypeTask {
		zap.L().Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsML.CommandAdd:
		cfg := getConfig(reqEvent)
		if cfg != nil && helper.IsMine(svcCFG.IDs, svcCFG.Labels, cfg.ID, cfg.Labels) {
			tasksStore.Add(*cfg)
		}

	case rsML.CommandRemove:
		cfg := getConfig(reqEvent)
		if cfg != nil {
			tasksStore.Remove(cfg.ID)
		}

	case rsML.CommandUnloadAll:
		tasksStore.RemoveAll()

	case rsML.CommandReload:
		cfg := getConfig(reqEvent)
		if cfg != nil {
			tasksStore.Remove(cfg.ID)
			if helper.IsMine(svcCFG.IDs, svcCFG.Labels, cfg.ID, cfg.Labels) {
				tasksStore.Add(*cfg)
			}
		}

	default:
		zap.L().Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func getConfig(reqEvent *rsML.Event) *taskML.Config {
	cfg := &taskML.Config{}
	err := reqEvent.ToStruct(cfg)
	if err != nil {
		zap.L().Error("Error on data conversion", zap.Error(err))
		return nil
	}
	return cfg
}

package handler

import (
	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

const (
	serviceMessageQueueLimit = 100
	serviceMessageQueueName  = "service_listener_handler"
)

// Config of handler service
type Config struct {
	IDs    []string
	Labels cmap.CustomStringMap
}

var (
	serviceQueue *queueUtils.Queue
	svcCFG       *Config
)

// Init handler service listener
func Init(config cmap.CustomMap) error {
	svcCFG = &Config{}
	err := utils.MapToStruct(utils.TagNameNone, config, svcCFG)
	if err != nil {
		return err
	}

	serviceQueue = queueUtils.New(serviceMessageQueueName, serviceMessageQueueLimit, postProcessServiceEvent, 1)

	// on message receive add it in to our local queue
	_, err = mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicServiceNotifyHandler), onServiceEvent)
	if err != nil {
		return err
	}

	err = initMessageListener()
	if err != nil {
		return err
	}

	// load handlers
	reqEvent := rsML.Event{
		Type:    rsML.TypeNotifyHandler,
		Command: rsML.CommandLoadAll,
	}
	topicResourceServer := mcbus.FormatTopic(mcbus.TopicServiceResourceServer)
	return mcbus.Publish(topicResourceServer, reqEvent)
}

// Close the service listener
func Close() {
	UnloadAll()
	serviceQueue.Close()
	closeMessageListener()
}

func onServiceEvent(event *busML.BusData) {
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
	status := serviceQueue.Produce(reqEvent)
	if !status {
		zap.L().Warn("Failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// postProcessServiceEvent from the queue
func postProcessServiceEvent(event interface{}) {
	reqEvent := event.(*rsML.Event)
	zap.L().Debug("Processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsML.TypeNotifyHandler {
		zap.L().Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsML.CommandStart:
		cfg := getConfig(reqEvent)
		if cfg != nil && helper.IsMine(svcCFG.IDs, svcCFG.Labels, cfg.ID, cfg.Labels) {
			err := Start(cfg)
			if err != nil {
				zap.L().Error("error on starting a handler", zap.Error(err), zap.String("handler", cfg.ID))
			}
		}

	case rsML.CommandStop:
		if reqEvent.ID != "" {
			err := Stop(reqEvent.ID)
			if err != nil {
				zap.L().Error("error on stopping a service", zap.Error(err))
			}
			return
		}
		cfg := getConfig(reqEvent)
		if cfg != nil {
			err := Stop(cfg.ID)
			if err != nil {
				zap.L().Error("error on stopping a service", zap.Error(err))
			}
		}

	case rsML.CommandReload:
		cfg := getConfig(reqEvent)
		if cfg != nil && helper.IsMine(svcCFG.IDs, svcCFG.Labels, cfg.ID, cfg.Labels) {
			err := Reload(cfg)
			if err != nil {
				zap.L().Error("error on reload a service", zap.Error(err))
			}
		}

	case rsML.CommandUnloadAll:
		UnloadAll()

	default:
		zap.L().Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func getConfig(reqEvent *rsML.Event) *handlerML.Config {
	cfg := &handlerML.Config{}
	err := reqEvent.ToStruct(cfg)
	if err != nil {
		zap.L().Error("Error on data conversion", zap.Error(err))
		return nil
	}
	return cfg
}

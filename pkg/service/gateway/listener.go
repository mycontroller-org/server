package service

import (
	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	rsml "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	helper "github.com/mycontroller-org/backend/v2/pkg/utils/filter_sort"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

var (
	eventQueue *queueUtils.Queue
	queueSize  = int(50)
	workers    = int(1)
	cfg        *Config
)

// Config of gateway service
type Config struct {
	IDs    []string
	Labels cmap.CustomStringMap
}

// Init starts resource server listener
func Init(config cmap.CustomMap) error {
	cfg = &Config{}
	err := utils.MapToStruct(utils.TagNameNone, config, cfg)
	if err != nil {
		return err
	}

	eventQueue = queueUtils.New("gateway_service", queueSize, processEvent, workers)

	// on event receive add it in to our local queue
	topic := mcbus.FormatTopic(mcbus.TopicServiceGateway)
	_, err = mcbus.Subscribe(topic, onEvent)
	if err != nil {
		return err
	}

	// load gateways
	reqEvent := rsml.Event{
		Type:    rsml.TypeGateway,
		Command: rsml.CommandLoadAll,
	}
	topicResourceServer := mcbus.FormatTopic(mcbus.TopicServiceResourceServer)
	return mcbus.Publish(topicResourceServer, reqEvent)
}

// Close the service
func Close() {
	UnloadAll()
	eventQueue.Close()
}

func onEvent(event *busML.BusData) {
	reqEvent := &rsml.Event{}
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
	status := eventQueue.Produce(reqEvent)
	if !status {
		zap.L().Warn("Failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processEvent from the queue
func processEvent(event interface{}) {
	reqEvent := event.(*rsml.Event)
	zap.L().Debug("Processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsml.TypeGateway {
		zap.L().Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsml.CommandStart:
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil && helper.IsMine(cfg.IDs, cfg.Labels, gwCfg.ID, gwCfg.Labels) {
			err := Start(gwCfg)
			if err != nil {
				zap.L().Error("error on starting a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
		}

	case rsml.CommandStop:
		if reqEvent.ID != "" {
			err := Stop(reqEvent.ID)
			if err != nil {
				zap.L().Error("error on stopping a service", zap.Error(err), zap.String("id", reqEvent.ID))
			}
			return
		}
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil {
			err := Stop(gwCfg.ID)
			if err != nil {
				zap.L().Error("error on stopping a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
		}

	case rsml.CommandReload:
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil {
			err := Stop(gwCfg.ID)
			if err != nil {
				zap.L().Error("error on stopping a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
			if helper.IsMine(cfg.IDs, cfg.Labels, gwCfg.ID, gwCfg.Labels) {
				err := Start(gwCfg)
				if err != nil {
					zap.L().Error("error on starting a service", zap.Error(err), zap.String("id", gwCfg.ID))
				}
			}
		}

	case rsml.CommandUnloadAll:
		UnloadAll()

	default:
		zap.L().Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func getGatewayConfig(reqEvent *rsml.Event) *gwml.Config {
	gwCfg := &gwml.Config{}
	err := reqEvent.ToStruct(gwCfg)
	if err != nil {
		zap.L().Error("Error on data conversion", zap.Error(err))
		return nil
	}
	return gwCfg
}

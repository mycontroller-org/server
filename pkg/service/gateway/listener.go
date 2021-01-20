package service

import (
	q "github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	"github.com/mycontroller-org/backend/v2/pkg/model/event"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	rsml "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
)

var (
	eventQueue *q.BoundedQueue
	queueSize  = int(50)
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

	eventQueue = utils.GetQueue("gateway_service", queueSize)

	// on event receive add it in to our local queue
	topic := mcbus.FormatTopic(mcbus.TopicServiceGateway)
	_, err = mcbus.Subscribe(topic, onEvent)
	if err != nil {
		return err
	}

	eventQueue.StartConsumers(1, processEvent)
	// load gateways
	reqEvent := rsml.Event{
		Type:    rsml.TypeGateway,
		Command: rsml.CommandLoadAll,
	}
	topicResourceServer := mcbus.FormatTopic(mcbus.TopicResourceServer)
	return mcbus.Publish(topicResourceServer, reqEvent)
}

// Close the service
func Close() {
	UnloadAll()
	eventQueue.Stop()
}

func onEvent(event *event.Event) {
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
		if gwCfg != nil && isMine(gwCfg) {
			Start(gwCfg)
		}

	case rsml.CommandStop:
		if reqEvent.ID != "" {
			Stop(reqEvent.ID)
			return
		}
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil {
			Stop(gwCfg.ID)
		}

	case rsml.CommandReload:
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil && isMine(gwCfg) {
			Reload(gwCfg)
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

func isMine(gwCfg *gwml.Config) bool {
	if len(cfg.IDs) == 0 {
		if len(cfg.Labels) == 0 {
			return true
		}
		for key, value := range cfg.Labels {
			receivedValue, found := gwCfg.Labels[key]
			if !found {
				return false
			}
			if value != receivedValue {
				return false
			}
		}
		return true
	}

	for _, id := range cfg.IDs {
		if id == gwCfg.ID {
			return true
		}
	}
	return false

}

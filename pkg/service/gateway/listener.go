package service

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	busML "github.com/mycontroller-org/server/v2/pkg/model/bus"
	rsML "github.com/mycontroller-org/server/v2/pkg/model/resource_service"
	sfML "github.com/mycontroller-org/server/v2/pkg/model/service_filter"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	gwType "github.com/mycontroller-org/server/v2/plugin/gateway/type"
	"go.uber.org/zap"
)

var (
	eventQueue *queueUtils.Queue
	queueSize  = int(50)
	workers    = int(1)
	svcFilter  *sfML.ServiceFilter
)

// Start starts resource server listener
func Start(filter *sfML.ServiceFilter) error {
	svcFilter = filter
	if svcFilter.Disabled {
		zap.L().Info("gateway service disabled")
		return nil
	}

	if svcFilter.HasFilter() {
		zap.L().Info("gateway service filter config", zap.Any("filter", svcFilter))
	} else {
		zap.L().Debug("there is no filter applied to gateway service")
	}

	eventQueue = queueUtils.New("gateway_service", queueSize, processEvent, workers)

	// on event receive add it in to our local queue
	topic := mcbus.FormatTopic(mcbus.TopicServiceGateway)
	_, err := mcbus.Subscribe(topic, onEvent)
	if err != nil {
		return err
	}

	// load gateways
	reqEvent := rsML.ServiceEvent{
		Type:    rsML.TypeGateway,
		Command: rsML.CommandLoadAll,
	}
	topicResourceServer := mcbus.FormatTopic(mcbus.TopicServiceResourceServer)
	return mcbus.Publish(topicResourceServer, reqEvent)
}

// Close the service
func Close() {
	if svcFilter.Disabled {
		return
	}
	UnloadAll()
	eventQueue.Close()
}

func onEvent(event *busML.BusData) {
	reqEvent := &rsML.ServiceEvent{}
	err := event.LoadData(reqEvent)
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
	reqEvent := event.(*rsML.ServiceEvent)
	zap.L().Debug("Processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsML.TypeGateway {
		zap.L().Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsML.CommandStart:
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil && helper.IsMine(svcFilter, gwCfg.Provider.GetString(model.KeyType), gwCfg.ID, gwCfg.Labels) {
			err := StartGW(gwCfg)
			if err != nil {
				zap.L().Error("error on starting a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
		}

	case rsML.CommandStop:
		if reqEvent.ID != "" {
			err := StopGW(reqEvent.ID)
			if err != nil {
				zap.L().Error("error on stopping a service", zap.Error(err), zap.String("id", reqEvent.ID))
			}
			return
		}
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil {
			err := StopGW(gwCfg.ID)
			if err != nil {
				zap.L().Error("error on stopping a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
		}

	case rsML.CommandReload:
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil {
			err := StopGW(gwCfg.ID)
			if err != nil {
				zap.L().Error("error on stopping a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
			if helper.IsMine(svcFilter, gwCfg.Provider.GetString(model.KeyType), gwCfg.ID, gwCfg.Labels) {
				err := StartGW(gwCfg)
				if err != nil {
					zap.L().Error("error on starting a service", zap.Error(err), zap.String("id", gwCfg.ID))
				}
			}
		}

	case rsML.CommandUnloadAll:
		UnloadAll()

	default:
		zap.L().Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func getGatewayConfig(reqEvent *rsML.ServiceEvent) *gwType.Config {
	cfg := &gwType.Config{}
	err := reqEvent.LoadData(cfg)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil
	}
	return cfg
}

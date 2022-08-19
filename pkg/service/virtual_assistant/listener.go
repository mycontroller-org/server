package service

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	vaTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_assistant"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	"go.uber.org/zap"
)

var (
	eventQueue *queueUtils.Queue
	queueSize  = int(50)
	workers    = int(1)
	svcFilter  *sfTY.ServiceFilter
)

// Start starts resource server listener
func Start(filter *sfTY.ServiceFilter) error {
	svcFilter = filter
	if svcFilter.Disabled {
		zap.L().Info("virtual assistant service disabled")
		return nil
	}

	if svcFilter.HasFilter() {
		zap.L().Info("virtual assistant service filter config", zap.Any("filter", svcFilter))
	} else {
		zap.L().Debug("there is no filter applied to virtual assistant service")
	}

	eventQueue = queueUtils.New("virtual_assistant_service", queueSize, processEvent, workers)

	// on event receive add it in to our local queue
	topic := mcbus.FormatTopic(mcbus.TopicServiceVirtualAssistant)
	_, err := mcbus.Subscribe(topic, onEvent)
	if err != nil {
		return err
	}

	// load virtual assistants
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeVirtualAssistant,
		Command: rsTY.CommandLoadAll,
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

func onEvent(event *busTY.BusData) {
	reqEvent := &rsTY.ServiceEvent{}
	err := event.LoadData(reqEvent)
	if err != nil {
		zap.L().Warn("failed to convert to target type", zap.Error(err))
		return
	}
	if reqEvent.Type == "" {
		zap.L().Warn("received an empty event", zap.Any("event", event))
		return
	}
	zap.L().Debug("event added into processing queue", zap.Any("event", reqEvent))
	status := eventQueue.Produce(reqEvent)
	if !status {
		zap.L().Warn("failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processEvent from the queue
func processEvent(event interface{}) {
	reqEvent := event.(*rsTY.ServiceEvent)
	zap.L().Debug("processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsTY.TypeVirtualAssistant {
		zap.L().Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsTY.CommandAdd:
		vaCfg := getConfig(reqEvent)
		if vaCfg != nil && helper.IsMine(svcFilter, vaCfg.ProviderType, vaCfg.ID, vaCfg.Labels) {
			err := StartAssistant(vaCfg)
			if err != nil {
				zap.L().Error("error on starting a service", zap.Error(err), zap.String("id", vaCfg.ID))
			}
		}

	case rsTY.CommandRemove:
		if reqEvent.ID != "" {
			err := StopAssistant(reqEvent.ID)
			if err != nil {
				zap.L().Error("error on stopping a service", zap.Error(err), zap.String("id", reqEvent.ID))
			}
			return
		}
		gwCfg := getConfig(reqEvent)
		if gwCfg != nil {
			err := StopAssistant(gwCfg.ID)
			if err != nil {
				zap.L().Error("error on stopping a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
		}

	case rsTY.CommandReload:
		gwCfg := getConfig(reqEvent)
		if gwCfg != nil {
			err := StopAssistant(gwCfg.ID)
			if err != nil {
				zap.L().Error("error on stopping a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
			if helper.IsMine(svcFilter, gwCfg.ProviderType, gwCfg.ID, gwCfg.Labels) {
				err := StartAssistant(gwCfg)
				if err != nil {
					zap.L().Error("error on starting a service", zap.Error(err), zap.String("id", gwCfg.ID))
				}
			}
		}

	case rsTY.CommandUnloadAll:
		UnloadAll()

	default:
		zap.L().Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func getConfig(reqEvent *rsTY.ServiceEvent) *vaTY.Config {
	cfg := &vaTY.Config{}
	err := reqEvent.LoadData(cfg)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil
	}
	return cfg
}

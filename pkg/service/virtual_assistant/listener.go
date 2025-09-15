package service

import (
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	vaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/types"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	"go.uber.org/zap"
)

func (svc *VirtualAssistantService) onEvent(event *busTY.BusData) {
	reqEvent := &rsTY.ServiceEvent{}
	err := event.LoadData(reqEvent)
	if err != nil {
		svc.logger.Warn("failed to convert to target type", zap.Error(err))
		return
	}
	if reqEvent.Type == "" {
		svc.logger.Warn("received an empty event", zap.Any("event", event))
		return
	}
	svc.logger.Debug("event added into processing queue", zap.Any("event", reqEvent))
	status := svc.eventsQueue.Produce(reqEvent)
	if !status {
		svc.logger.Warn("failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processEvent from the queue
func (svc *VirtualAssistantService) processEvent(event interface{}) error {
	reqEvent := event.(*rsTY.ServiceEvent)
	svc.logger.Debug("processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsTY.TypeVirtualAssistant {
		svc.logger.Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsTY.CommandAdd:
		vaCfg := svc.getConfig(reqEvent)
		if vaCfg != nil && helper.IsMine(svc.filter, vaCfg.ProviderType, vaCfg.ID, vaCfg.Labels) {
			err := svc.startAssistant(vaCfg)
			if err != nil {
				svc.logger.Error("error on starting a service", zap.Error(err), zap.String("id", vaCfg.ID))
			}
		}

	case rsTY.CommandRemove:
		if reqEvent.ID != "" {
			err := svc.stopAssistant(reqEvent.ID)
			if err != nil {
				svc.logger.Error("error on stopping a service", zap.Error(err), zap.String("id", reqEvent.ID))
			}
			return nil
		}
		gwCfg := svc.getConfig(reqEvent)
		if gwCfg != nil {
			err := svc.stopAssistant(gwCfg.ID)
			if err != nil {
				svc.logger.Error("error on stopping a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
		}

	case rsTY.CommandReload:
		gwCfg := svc.getConfig(reqEvent)
		if gwCfg != nil {
			err := svc.stopAssistant(gwCfg.ID)
			if err != nil {
				svc.logger.Error("error on stopping a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
			if helper.IsMine(svc.filter, gwCfg.ProviderType, gwCfg.ID, gwCfg.Labels) {
				err := svc.startAssistant(gwCfg)
				if err != nil {
					svc.logger.Error("error on starting a service", zap.Error(err), zap.String("id", gwCfg.ID))
				}
			}
		}

	case rsTY.CommandUnloadAll:
		svc.unloadAll()

	default:
		svc.logger.Warn("unsupported command", zap.Any("event", reqEvent))
	}
	return nil
}

func (svc *VirtualAssistantService) getConfig(reqEvent *rsTY.ServiceEvent) *vaTY.Config {
	cfg := &vaTY.Config{}
	err := reqEvent.LoadData(cfg)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil
	}
	return cfg
}

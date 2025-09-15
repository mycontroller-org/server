package service

import (
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

// Start starts resource server listener
func (svc *GatewayService) Start() error {
	if svc.filter.Disabled {
		svc.logger.Info("gateway service disabled")
		return nil
	}

	if svc.filter.HasFilter() {
		svc.logger.Info("gateway service filter config", zap.Any("filter", svc.filter))
	} else {
		svc.logger.Debug("there is no filter applied to gateway service")
	}

	// on event receive add it in to our local queue
	id, err := svc.bus.Subscribe(svc.eventsQueue.Topic, svc.onEvent)
	if err != nil {
		return err
	}
	svc.eventsQueue.SubscriptionId = id

	// load gateways
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeGateway,
		Command: rsTY.CommandLoadAll,
	}
	return svc.bus.Publish(topic.TopicServiceResourceServer, reqEvent)
}

// Close the service
func (svc *GatewayService) Close() error {
	if svc.filter.Disabled {
		return nil
	}
	svc.unloadAll()
	svc.eventsQueue.Close()
	return nil
}

func (svc *GatewayService) onEvent(event *busTY.BusData) {
	reqEvent := &rsTY.ServiceEvent{}
	err := event.LoadData(reqEvent)
	if err != nil {
		svc.logger.Warn("Failed to convert to target type", zap.Error(err))
		return
	}
	if reqEvent.Type == "" {
		svc.logger.Warn("received an empty event", zap.Any("event", event))
		return
	}
	svc.logger.Debug("Event added into processing queue", zap.Any("event", reqEvent))
	status := svc.eventsQueue.Produce(reqEvent)
	if !status {
		svc.logger.Warn("Failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processEvent from the queue
func (svc *GatewayService) processEvent(event interface{}) error {
	reqEvent := event.(*rsTY.ServiceEvent)
	svc.logger.Debug("Processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsTY.TypeGateway {
		svc.logger.Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsTY.CommandStart:
		gwCfg := svc.getGatewayConfig(reqEvent)
		if gwCfg != nil && helper.IsMine(svc.filter, gwCfg.Provider.GetString(types.KeyType), gwCfg.ID, gwCfg.Labels) {
			err := svc.startGW(gwCfg)
			if err != nil {
				svc.logger.Error("error on starting a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
		}

	case rsTY.CommandStop:
		if reqEvent.ID != "" {
			err := svc.stopGW(reqEvent.ID)
			if err != nil {
				svc.logger.Error("error on stopping a service", zap.Error(err), zap.String("id", reqEvent.ID))
			}
			return nil
		}
		gwCfg := svc.getGatewayConfig(reqEvent)
		if gwCfg != nil {
			err := svc.stopGW(gwCfg.ID)
			if err != nil {
				svc.logger.Error("error on stopping a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
		}

	case rsTY.CommandReload:
		gwCfg := svc.getGatewayConfig(reqEvent)
		if gwCfg != nil {
			err := svc.stopGW(gwCfg.ID)
			if err != nil {
				svc.logger.Error("error on stopping a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
			if helper.IsMine(svc.filter, gwCfg.Provider.GetString(types.KeyType), gwCfg.ID, gwCfg.Labels) {
				err := svc.startGW(gwCfg)
				if err != nil {
					svc.logger.Error("error on starting a service", zap.Error(err), zap.String("id", gwCfg.ID))
				}
			}
		}

	case rsTY.CommandUnloadAll:
		svc.unloadAll()

	case rsTY.CommandGetSleepingQueue:
		svc.processSleepingQueueRequest(reqEvent)

	case rsTY.CommandClearSleepingQueue:
		svc.clearSleepingQueue(reqEvent)

	default:
		svc.logger.Warn("unsupported command", zap.Any("event", reqEvent))
	}
	return nil
}

func (svc *GatewayService) getGatewayConfig(reqEvent *rsTY.ServiceEvent) *gwTY.Config {
	cfg := &gwTY.Config{}
	err := reqEvent.LoadData(cfg)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil
	}
	return cfg
}

// clear sleeping queue messages
func (svc *GatewayService) clearSleepingQueue(reqEvent *rsTY.ServiceEvent) {
	ids := make(map[string]interface{})
	err := reqEvent.LoadData(&ids)
	if err != nil {
		svc.logger.Error("error on parsing input", zap.Error(err), zap.Any("input", reqEvent))
		return
	}
	gatewayID := utils.GetMapValueString(ids, types.KeyGatewayID, "")
	nodeID := utils.GetMapValueString(ids, types.KeyNodeID, "")
	if gatewayID == "" {
		return
	}
	if nodeID != "" {
		svc.clearNodeSleepingQueue(gatewayID, nodeID)
	} else {
		svc.clearGatewaySleepingQueue(gatewayID)
	}
}

// process request from server and sends the available queue message details
func (svc *GatewayService) processSleepingQueueRequest(reqEvent *rsTY.ServiceEvent) {
	ids := make(map[string]interface{})
	err := reqEvent.LoadData(&ids)
	if err != nil {
		svc.logger.Error("error on parsing input", zap.Error(err), zap.Any("input", reqEvent))
		return
	}
	gatewayID := utils.GetMapValueString(ids, types.KeyGatewayID, "")
	nodeID := utils.GetMapValueString(ids, types.KeyNodeID, "")
	if gatewayID == "" {
		return
	}

	resEvent := &rsTY.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	messagesAvailable := false
	if nodeID != "" {
		receivedMessages := svc.getNodeSleepingQueue(gatewayID, nodeID)
		if receivedMessages != nil {
			resEvent.SetData(receivedMessages)
			messagesAvailable = true
		}
	} else {
		receivedMessages := svc.getGatewaySleepingQueue(gatewayID)
		if receivedMessages != nil {
			resEvent.SetData(receivedMessages)
			messagesAvailable = true
		}
	}

	if messagesAvailable {
		err = svc.postResponse(reqEvent.ReplyTopic, resEvent)
		if err != nil {
			svc.logger.Error("error on sending response", zap.Error(err), zap.Any("request", reqEvent))
		}
	}
}

// post response to a topic
func (svc *GatewayService) postResponse(topic string, response *rsTY.ServiceEvent) error {
	if topic == "" {
		return nil
	}
	return svc.bus.Publish(topic, response)
}

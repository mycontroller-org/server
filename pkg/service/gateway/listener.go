package service

import (
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	busTY "github.com/mycontroller-org/server/v2/pkg/types/bus"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	sfTY "github.com/mycontroller-org/server/v2/pkg/types/service_filter"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	helper "github.com/mycontroller-org/server/v2/pkg/utils/filter_sort"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/type"
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
	reqEvent := rsTY.ServiceEvent{
		Type:    rsTY.TypeGateway,
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
	reqEvent := event.(*rsTY.ServiceEvent)
	zap.L().Debug("Processing a request", zap.Any("event", reqEvent))

	if reqEvent.Type != rsTY.TypeGateway {
		zap.L().Warn("unsupported event type", zap.Any("event", reqEvent))
	}

	switch reqEvent.Command {
	case rsTY.CommandStart:
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil && helper.IsMine(svcFilter, gwCfg.Provider.GetString(types.KeyType), gwCfg.ID, gwCfg.Labels) {
			err := StartGW(gwCfg)
			if err != nil {
				zap.L().Error("error on starting a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
		}

	case rsTY.CommandStop:
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

	case rsTY.CommandReload:
		gwCfg := getGatewayConfig(reqEvent)
		if gwCfg != nil {
			err := StopGW(gwCfg.ID)
			if err != nil {
				zap.L().Error("error on stopping a service", zap.Error(err), zap.String("id", gwCfg.ID))
			}
			if helper.IsMine(svcFilter, gwCfg.Provider.GetString(types.KeyType), gwCfg.ID, gwCfg.Labels) {
				err := StartGW(gwCfg)
				if err != nil {
					zap.L().Error("error on starting a service", zap.Error(err), zap.String("id", gwCfg.ID))
				}
			}
		}

	case rsTY.CommandUnloadAll:
		UnloadAll()

	case rsTY.CommandGetSleepingQueue:
		processSleepingQueueRequest(reqEvent)

	case rsTY.CommandClearSleepingQueue:
		clearSleepingQueue(reqEvent)

	default:
		zap.L().Warn("unsupported command", zap.Any("event", reqEvent))
	}
}

func getGatewayConfig(reqEvent *rsTY.ServiceEvent) *gwTY.Config {
	cfg := &gwTY.Config{}
	err := reqEvent.LoadData(cfg)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil
	}
	return cfg
}

// clear sleeping queue messages
func clearSleepingQueue(reqEvent *rsTY.ServiceEvent) {
	ids := make(map[string]interface{})
	err := reqEvent.LoadData(&ids)
	if err != nil {
		zap.L().Error("error on parsing input", zap.Error(err), zap.Any("input", reqEvent))
		return
	}
	gatewayID := utils.GetMapValueString(ids, types.KeyGatewayID, "")
	nodeID := utils.GetMapValueString(ids, types.KeyNodeID, "")
	if gatewayID == "" {
		return
	}
	if nodeID != "" {
		clearNodeSleepingQueue(gatewayID, nodeID)
	} else {
		clearGatewaySleepingQueue(gatewayID)
	}
}

// process request from server and sends the available queue message details
func processSleepingQueueRequest(reqEvent *rsTY.ServiceEvent) {
	ids := make(map[string]interface{})
	err := reqEvent.LoadData(&ids)
	if err != nil {
		zap.L().Error("error on parsing input", zap.Error(err), zap.Any("input", reqEvent))
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

	var receivedMessages interface{}
	if nodeID != "" {
		receivedMessages = getNodeSleepingQueue(gatewayID, nodeID)
	} else {
		receivedMessages = getGatewaySleepingQueue(gatewayID)
	}

	if receivedMessages != nil {
		resEvent.SetData(receivedMessages)
		err = postResponse(reqEvent.ReplyTopic, resEvent)
		if err != nil {
			zap.L().Error("error on sending response", zap.Error(err), zap.Any("request", reqEvent))
		}
	}
}

// post response to a topic
func postResponse(topic string, response *rsTY.ServiceEvent) error {
	if topic == "" {
		return nil
	}
	return mcbus.Publish(topic, response)
}

package service

import (
	"errors"
	"fmt"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	gatewayAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	"github.com/mycontroller-org/backend/v2/pkg/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	busml "github.com/mycontroller-org/backend/v2/plugin/bus"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
	"go.uber.org/zap"
)

var (
	eventQueue *q.BoundedQueue
	queueSize  = int(1000)
)

// Init starts resource server listener
func Init() error {
	eventQueue = utils.GetQueue("resource_service", queueSize)

	// on event receive add it in to our local queue
	_, err := mcbus.Subscribe(mcbus.FormatTopic(mcbus.TopicResourceServer), onEvent)
	if err != nil {
		return err
	}

	eventQueue.StartConsumers(1, processEvent)
	return nil
}

// Close the service
func Close() {
	eventQueue.Stop()
}

func onEvent(event *busml.Event) {
	reqEvent := &rsModel.Event{}
	err := event.ToStruct(reqEvent)
	if err != nil {
		zap.L().Warn("Failed to convet to target type", zap.Error(err))
		return
	}

	if reqEvent == nil {
		zap.L().Warn("Received a nil event", zap.Any("event", event))
		return
	}
	zap.L().Debug("Event added into processing queue", zap.Any("event", reqEvent))
	status := eventQueue.Produce(reqEvent)
	if !status {
		zap.L().Warn("Failed to store the event into queue", zap.Any("event", reqEvent))
	}
}

// processEvent from the queue
func processEvent(item interface{}) {
	request := item.(*rsModel.Event)
	zap.L().Debug("Processing an event", zap.Any("event", request))

	switch rsModel.TypeGateway {
	case rsModel.TypeGateway:
		gatewayRequest(request)
	default:
		zap.L().Warn("unknown event type", zap.Any("event", request))
	}
}

func gatewayRequest(reqEvent *rsModel.Event) error {
	resEvent := &rsModel.Event{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}
	switch reqEvent.Command {
	case rsModel.CommandGet:
		data, err := getGateway(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		err = resEvent.SetData(data)
		if err != nil {
			return err
		}

	case rsModel.CommandUpdateState:
		err := updateGatewayState(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}

	case rsModel.CommandLoadAll:
		gatewayAPI.LoadAll()

	default:
		return errors.New("Unknown command")
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func postResponse(topic string, response *rsModel.Event) error {
	if topic == "" {
		return nil
	}
	return mcbus.Publish(topic, response)
}

func getLabelsFilter(labels cmap.CustomStringMap) []stgml.Filter {
	filters := make([]stgml.Filter, 0)
	for key, value := range labels {
		filter := stgml.Filter{Key: fmt.Sprintf("labels.%s", key), Operator: stgml.OperatorEqual, Value: value}
		filters = append(filters, filter)
	}
	return filters
}

func getGateway(request *rsModel.Event) (interface{}, error) {
	if request.ID != "" {
		gwConfig, err := gatewayAPI.GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return gwConfig, nil
	} else if len(request.Labels) > 0 {
		filters := getLabelsFilter(request.Labels)
		result, err := gatewayAPI.List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func updateGatewayState(reqEvent *rsModel.Event) error {
	if reqEvent.Data == nil {
		zap.L().Error("gateway state not supplied", zap.Any("event", reqEvent))
		return errors.New("gateway state not supplied")
	}
	state := model.State{}
	err := reqEvent.ToStruct(&state)
	if err != nil {
		return err
	}
	return gatewayAPI.SetState(reqEvent.ID, state)
}

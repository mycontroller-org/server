package resource

import (
	"errors"

	gatewayAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"go.uber.org/zap"
)

func gatewayService(reqEvent *rsModel.Event) error {
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
	state := &model.State{}
	err := reqEvent.ToStruct(state)
	if err != nil {
		return err
	}
	return gatewayAPI.SetState(reqEvent.ID, state)
}

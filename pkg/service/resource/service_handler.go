package resource

import (
	"errors"

	handlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/handler"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"go.uber.org/zap"
)

func handlerService(reqEvent *rsML.ServiceEvent) error {
	resEvent := &rsML.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsML.CommandGet:
		data, err := getHandler(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsML.CommandUpdateState:
		err := updateHandlerState(reqEvent)
		if err != nil {
			return err
		}

	case rsML.CommandLoadAll:
		handlerAPI.LoadAll()

	default:
		return errors.New("unknown command")
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getHandler(request *rsML.ServiceEvent) (interface{}, error) {
	if request.ID != "" {
		cfg, err := handlerAPI.GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if len(request.Labels) > 0 {
		filters := getLabelsFilter(request.Labels)
		result, err := handlerAPI.List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func updateHandlerState(reqEvent *rsML.ServiceEvent) error {
	if reqEvent.Data == nil {
		zap.L().Error("handler state not supplied", zap.Any("event", reqEvent))
		return errors.New("handler state not supplied")
	}
	state := &model.State{}
	err := reqEvent.LoadData(state)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return err
	}

	return handlerAPI.SetState(reqEvent.ID, state)
}

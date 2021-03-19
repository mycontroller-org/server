package resource

import (
	"errors"

	handlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/notify_handler"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"go.uber.org/zap"
)

func handlerService(reqEvent *rsModel.Event) error {
	resEvent := &rsModel.Event{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsModel.CommandGet:
		data, err := getHandler(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		err = resEvent.SetData(data)
		if err != nil {
			return err
		}

	case rsModel.CommandUpdateState:
		err := updateHandlerState(reqEvent)
		if err != nil {
			return err
		}

	case rsModel.CommandLoadAll:
		handlerAPI.LoadAll()

	default:
		return errors.New("unknown command")
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getHandler(request *rsModel.Event) (interface{}, error) {
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

func updateHandlerState(reqEvent *rsModel.Event) error {
	if reqEvent.Data == nil {
		zap.L().Error("handler state not supplied", zap.Any("event", reqEvent))
		return errors.New("handler state not supplied")
	}
	state := &model.State{}
	err := reqEvent.ToStruct(state)
	if err != nil {
		return err
	}
	return handlerAPI.SetState(reqEvent.ID, state)
}

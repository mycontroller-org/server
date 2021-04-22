package resource

import (
	"errors"
	"fmt"

	handlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/handler"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"go.uber.org/zap"
)

func handlerService(reqEvent *rsModel.ServiceEvent) error {
	resEvent := &rsModel.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsModel.CommandGet:
		data, err := getHandler(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

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

func getHandler(request *rsModel.ServiceEvent) (interface{}, error) {
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

func updateHandlerState(reqEvent *rsModel.ServiceEvent) error {
	if reqEvent.Data == nil {
		zap.L().Error("handler state not supplied", zap.Any("event", reqEvent))
		return errors.New("handler state not supplied")
	}
	state, ok := reqEvent.GetData().(model.State)
	if !ok {
		return fmt.Errorf("error on data conversion, receivedType: %T", reqEvent.GetData())
	}

	return handlerAPI.SetState(reqEvent.ID, &state)
}

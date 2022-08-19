package resource

import (
	"errors"

	vaAPI "github.com/mycontroller-org/server/v2/pkg/api/virtual_assistant"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"go.uber.org/zap"
)

func virtualAssistantService(reqEvent *rsTY.ServiceEvent) error {
	resEvent := &rsTY.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsTY.CommandGet:
		data, err := getVirtualAssistant(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsTY.CommandUpdateState:
		err := updateVirtualAssistantState(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}

	case rsTY.CommandLoadAll:
		vaAPI.LoadAll()

	case rsTY.CommandDisable:
		return disableVirtualAssistant(reqEvent)

	default:
		return errors.New("unknown command")
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getVirtualAssistant(request *rsTY.ServiceEvent) (interface{}, error) {
	if request.ID != "" {
		cfg, err := vaAPI.GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if len(request.Labels) > 0 {
		filters := getLabelsFilter(request.Labels)
		result, err := vaAPI.List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func updateVirtualAssistantState(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == "" {
		zap.L().Error("state not supplied", zap.Any("event", reqEvent))
		return errors.New("state not supplied")
	}

	state := &types.State{}
	err := reqEvent.LoadData(state)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return err
	}

	return vaAPI.SetState(reqEvent.ID, state)
}

func disableVirtualAssistant(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == "" {
		zap.L().Error("virtual assistant id not supplied", zap.Any("event", reqEvent))
		return errors.New("id not supplied")
	}

	id := ""
	err := reqEvent.LoadData(&id)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("reqEvent", reqEvent), zap.Error(err))
		return err
	}

	return vaAPI.Disable([]string{id})
}

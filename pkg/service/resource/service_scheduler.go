package resource

import (
	"errors"

	schedulerAPI "github.com/mycontroller-org/backend/v2/pkg/api/scheduler"
	rsML "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	schedulerML "github.com/mycontroller-org/backend/v2/pkg/model/scheduler"
	"go.uber.org/zap"
)

func schedulerService(reqEvent *rsML.Event) error {
	resEvent := &rsML.Event{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsML.CommandGet:
		data, err := getHandler(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		err = resEvent.SetData(data)
		if err != nil {
			return err
		}

	case rsML.CommandUpdateState:
		err := updateSchedulerState(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}

	case rsML.CommandLoadAll:
		schedulerAPI.LoadAll()

	case rsML.CommandDisable:
		return disableScheduler(reqEvent)

	default:
		return errors.New("Unknown command")
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getScheduler(request *rsML.Event) (interface{}, error) {
	if request.ID != "" {
		cfg, err := schedulerAPI.GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if len(request.Labels) > 0 {
		filters := getLabelsFilter(request.Labels)
		result, err := schedulerAPI.List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func updateSchedulerState(reqEvent *rsML.Event) error {
	if reqEvent.Data == nil {
		zap.L().Error("state not supplied", zap.Any("event", reqEvent))
		return errors.New("state not supplied")
	}
	state := &schedulerML.State{}
	err := reqEvent.ToStruct(state)
	if err != nil {
		return err
	}
	return schedulerAPI.SetState(reqEvent.ID, state)
}

func disableScheduler(reqEvent *rsML.Event) error {
	if reqEvent.Data == nil {
		zap.L().Error("scheduler id not supplied", zap.Any("event", reqEvent))
		return errors.New("id not supplied")
	}
	var id string
	err := reqEvent.ToStruct(&id)
	if err != nil {
		return err
	}
	return schedulerAPI.Disable([]string{id})

}

package resource

import (
	"errors"

	taskAPI "github.com/mycontroller-org/backend/v2/pkg/api/task"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/model/task"
	"go.uber.org/zap"
)

func taskService(reqEvent *rsModel.Event) error {
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
		err := updateTaskState(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}

	case rsModel.CommandLoadAll:
		taskAPI.LoadAll()

	default:
		return errors.New("Unknown command")
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getTask(request *rsModel.Event) (interface{}, error) {
	if request.ID != "" {
		cfg, err := taskAPI.GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if len(request.Labels) > 0 {
		filters := getLabelsFilter(request.Labels)
		result, err := taskAPI.List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func updateTaskState(reqEvent *rsModel.Event) error {
	if reqEvent.Data == nil {
		zap.L().Error("handler state not supplied", zap.Any("event", reqEvent))
		return errors.New("handler state not supplied")
	}
	state := &task.State{}
	err := reqEvent.ToStruct(state)
	if err != nil {
		return err
	}
	return taskAPI.SetState(reqEvent.ID, state)
}

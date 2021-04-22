package resource

import (
	"errors"
	"fmt"

	taskAPI "github.com/mycontroller-org/backend/v2/pkg/api/task"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"github.com/mycontroller-org/backend/v2/pkg/model/task"
	"go.uber.org/zap"
)

func taskService(reqEvent *rsModel.ServiceEvent) error {
	resEvent := &rsModel.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsModel.CommandGet:
		data, err := getTask(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsModel.CommandUpdateState:
		err := updateTaskState(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}

	case rsModel.CommandLoadAll:
		taskAPI.LoadAll()

	case rsModel.CommandDisable:
		return disableTask(reqEvent)

	default:
		return fmt.Errorf("unknown command: %s", reqEvent.Command)
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getTask(request *rsModel.ServiceEvent) (interface{}, error) {
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

func updateTaskState(reqEvent *rsModel.ServiceEvent) error {
	if reqEvent.Data == nil {
		zap.L().Error("handler state not supplied", zap.Any("event", reqEvent))
		return errors.New("handler state not supplied")
	}

	state, ok := reqEvent.GetData().(task.State)
	if !ok {
		return fmt.Errorf("error on data conversion, receivedType: %T", reqEvent.GetData())
	}

	return taskAPI.SetState(reqEvent.ID, &state)
}

func disableTask(reqEvent *rsModel.ServiceEvent) error {
	if reqEvent.Data == nil {
		zap.L().Error("Task id not supplied", zap.Any("event", reqEvent))
		return errors.New("id not supplied")
	}

	id, ok := reqEvent.GetData().(string)
	if !ok {
		return fmt.Errorf("error on data conversion, receivedType: %T", reqEvent.GetData())
	}
	return taskAPI.Disable([]string{id})
}

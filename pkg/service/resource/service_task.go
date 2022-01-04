package resource

import (
	"errors"
	"fmt"

	taskAPI "github.com/mycontroller-org/server/v2/pkg/api/task"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/types/task"
	"go.uber.org/zap"
)

func taskService(reqEvent *rsTY.ServiceEvent) error {
	resEvent := &rsTY.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsTY.CommandGet:
		data, err := getTask(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsTY.CommandUpdateState:
		err := updateTaskState(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}

	case rsTY.CommandLoadAll:
		taskAPI.LoadAll()

	case rsTY.CommandDisable:
		return disableTask(reqEvent)

	default:
		return fmt.Errorf("unknown command: %s", reqEvent.Command)
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getTask(request *rsTY.ServiceEvent) (interface{}, error) {
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

func updateTaskState(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == "" {
		zap.L().Error("handler state not supplied", zap.Any("event", reqEvent))
		return errors.New("handler state not supplied")
	}

	state := &task.State{}
	err := reqEvent.LoadData(state)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return err
	}

	return taskAPI.SetState(reqEvent.ID, state)
}

func disableTask(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == "" {
		zap.L().Error("task id not supplied", zap.Any("event", reqEvent))
		return errors.New("id not supplied")
	}

	id := ""
	err := reqEvent.LoadData(&id)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("reqEvent", reqEvent), zap.Error(err))
		return err
	}

	return taskAPI.Disable([]string{id})
}

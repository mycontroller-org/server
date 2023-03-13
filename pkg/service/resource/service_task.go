package resource

import (
	"errors"
	"fmt"

	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"github.com/mycontroller-org/server/v2/pkg/types/task"
	"go.uber.org/zap"
)

func (svc *ResourceService) taskService(reqEvent *rsTY.ServiceEvent) error {
	resEvent := &rsTY.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsTY.CommandGet:
		data, err := svc.getTask(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsTY.CommandUpdateState:
		err := svc.updateTaskState(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}

	case rsTY.CommandLoadAll:
		svc.api.Task().LoadAll()

	case rsTY.CommandEnable:
		return svc.enableOrDisableTask(reqEvent, true)

	case rsTY.CommandDisable:
		return svc.enableOrDisableTask(reqEvent, false)

	default:
		return fmt.Errorf("unknown command: %s", reqEvent.Command)
	}
	return svc.postResponse(reqEvent.ReplyTopic, resEvent)
}

func (svc *ResourceService) getTask(request *rsTY.ServiceEvent) (interface{}, error) {
	if request.ID != "" {
		cfg, err := svc.api.Task().GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if len(request.Labels) > 0 {
		filters := svc.getLabelsFilter(request.Labels)
		result, err := svc.api.Task().List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func (svc *ResourceService) updateTaskState(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == "" {
		svc.logger.Error("handler state not supplied", zap.Any("event", reqEvent))
		return errors.New("handler state not supplied")
	}

	state := &task.State{}
	err := reqEvent.LoadData(state)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return err
	}

	return svc.api.Task().SetState(reqEvent.ID, state)
}

func (svc *ResourceService) enableOrDisableTask(reqEvent *rsTY.ServiceEvent, enable bool) error {
	if reqEvent.Data == "" {
		svc.logger.Error("task id not supplied", zap.Any("event", reqEvent))
		return errors.New("id not supplied")
	}

	id := ""
	err := reqEvent.LoadData(&id)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("reqEvent", reqEvent), zap.Error(err))
		return err
	}

	if enable {
		return svc.api.Task().Enable([]string{id})
	}
	return svc.api.Task().Disable([]string{id})
}

package resource

import (
	"errors"

	scheduleAPI "github.com/mycontroller-org/server/v2/pkg/api/schedule"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	"go.uber.org/zap"
)

func schedulerService(reqEvent *rsTY.ServiceEvent) error {
	resEvent := &rsTY.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsTY.CommandGet:
		data, err := getScheduler(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsTY.CommandUpdateState:
		err := updateSchedulerState(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}

	case rsTY.CommandLoadAll:
		scheduleAPI.LoadAll()

	case rsTY.CommandDisable:
		return disableScheduler(reqEvent)

	default:
		return errors.New("unknown command")
	}
	return postResponse(reqEvent.ReplyTopic, resEvent)
}

func getScheduler(request *rsTY.ServiceEvent) (interface{}, error) {
	if request.ID != "" {
		cfg, err := scheduleAPI.GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if len(request.Labels) > 0 {
		filters := getLabelsFilter(request.Labels)
		result, err := scheduleAPI.List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func updateSchedulerState(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == nil {
		zap.L().Error("state not supplied", zap.Any("event", reqEvent))
		return errors.New("state not supplied")
	}

	state := &scheduleTY.State{}
	err := reqEvent.LoadData(state)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return err
	}

	return scheduleAPI.SetState(reqEvent.ID, state)
}

func disableScheduler(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == nil {
		zap.L().Error("scheduler id not supplied", zap.Any("event", reqEvent))
		return errors.New("id not supplied")
	}

	id := ""
	err := reqEvent.LoadData(&id)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("reqEvent", reqEvent), zap.Error(err))
		return err
	}

	return scheduleAPI.Disable([]string{id})
}

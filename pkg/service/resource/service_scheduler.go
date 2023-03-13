package resource

import (
	"errors"

	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	"go.uber.org/zap"
)

func (svc *ResourceService) schedulerService(reqEvent *rsTY.ServiceEvent) error {
	resEvent := &rsTY.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsTY.CommandGet:
		data, err := svc.getScheduler(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsTY.CommandUpdateState:
		err := svc.updateSchedulerState(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}

	case rsTY.CommandLoadAll:
		svc.api.Schedule().LoadAll()

	case rsTY.CommandDisable:
		return svc.disableScheduler(reqEvent)

	default:
		return errors.New("unknown command")
	}
	return svc.postResponse(reqEvent.ReplyTopic, resEvent)
}

func (svc *ResourceService) getScheduler(request *rsTY.ServiceEvent) (interface{}, error) {
	if request.ID != "" {
		cfg, err := svc.api.Schedule().GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if len(request.Labels) > 0 {
		filters := svc.getLabelsFilter(request.Labels)
		result, err := svc.api.Schedule().List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func (svc *ResourceService) updateSchedulerState(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == "" {
		svc.logger.Error("state not supplied", zap.Any("event", reqEvent))
		return errors.New("state not supplied")
	}

	state := &schedulerTY.State{}
	err := reqEvent.LoadData(state)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return err
	}

	return svc.api.Schedule().SetState(reqEvent.ID, state)
}

func (svc *ResourceService) disableScheduler(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == "" {
		svc.logger.Error("scheduler id not supplied", zap.Any("event", reqEvent))
		return errors.New("id not supplied")
	}

	id := ""
	err := reqEvent.LoadData(&id)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("reqEvent", reqEvent), zap.Error(err))
		return err
	}

	return svc.api.Schedule().Disable([]string{id})
}

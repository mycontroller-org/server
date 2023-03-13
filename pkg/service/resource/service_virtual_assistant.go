package resource

import (
	"errors"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"go.uber.org/zap"
)

func (svc *ResourceService) virtualAssistantService(reqEvent *rsTY.ServiceEvent) error {
	resEvent := &rsTY.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsTY.CommandGet:
		data, err := svc.getVirtualAssistant(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsTY.CommandUpdateState:
		err := svc.updateVirtualAssistantState(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}

	case rsTY.CommandLoadAll:
		svc.api.VirtualAssistant().LoadAll()

	case rsTY.CommandDisable:
		return svc.disableVirtualAssistant(reqEvent)

	default:
		return errors.New("unknown command")
	}
	return svc.postResponse(reqEvent.ReplyTopic, resEvent)
}

func (svc *ResourceService) getVirtualAssistant(request *rsTY.ServiceEvent) (interface{}, error) {
	if request.ID != "" {
		cfg, err := svc.api.VirtualAssistant().GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if len(request.Labels) > 0 {
		filters := svc.getLabelsFilter(request.Labels)
		result, err := svc.api.VirtualAssistant().List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func (svc *ResourceService) updateVirtualAssistantState(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == "" {
		svc.logger.Error("state not supplied", zap.Any("event", reqEvent))
		return errors.New("state not supplied")
	}

	state := &types.State{}
	err := reqEvent.LoadData(state)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return err
	}

	return svc.api.VirtualAssistant().SetState(reqEvent.ID, state)
}

func (svc *ResourceService) disableVirtualAssistant(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == "" {
		svc.logger.Error("virtual assistant id not supplied", zap.Any("event", reqEvent))
		return errors.New("id not supplied")
	}

	id := ""
	err := reqEvent.LoadData(&id)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("reqEvent", reqEvent), zap.Error(err))
		return err
	}

	return svc.api.VirtualAssistant().Disable([]string{id})
}

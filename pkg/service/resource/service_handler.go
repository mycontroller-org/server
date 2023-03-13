package resource

import (
	"errors"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"go.uber.org/zap"
)

func (svc *ResourceService) handlerService(reqEvent *rsTY.ServiceEvent) error {
	resEvent := &rsTY.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsTY.CommandGet:
		data, err := svc.getHandler(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	case rsTY.CommandUpdateState:
		err := svc.updateHandlerState(reqEvent)
		if err != nil {
			return err
		}

	case rsTY.CommandLoadAll:
		svc.api.Handler().LoadAll()

	default:
		return errors.New("unknown command")
	}
	return svc.postResponse(reqEvent.ReplyTopic, resEvent)
}

func (svc *ResourceService) getHandler(request *rsTY.ServiceEvent) (interface{}, error) {
	if request.ID != "" {
		cfg, err := svc.api.Handler().GetByID(request.ID)
		if err != nil {
			return nil, err
		}
		return cfg, nil
	} else if len(request.Labels) > 0 {
		filters := svc.getLabelsFilter(request.Labels)
		result, err := svc.api.Handler().List(filters, nil)
		if err != nil {
			return nil, err
		}
		return result.Data, nil
	}
	return nil, errors.New("filter not supplied")
}

func (svc *ResourceService) updateHandlerState(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Data == "" {
		svc.logger.Error("handler state not supplied", zap.Any("event", reqEvent))
		return errors.New("handler state not supplied")
	}
	state := &types.State{}
	err := reqEvent.LoadData(state)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return err
	}

	return svc.api.Handler().SetState(reqEvent.ID, state)
}

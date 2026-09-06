package resource

import (
	"errors"

	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	"go.uber.org/zap"
)

func (svc *ResourceService) dataRepositoryService(reqEvent *rsTY.ServiceEvent) error {
	resEvent := &rsTY.ServiceEvent{
		Type:    reqEvent.Type,
		Command: reqEvent.ReplyCommand,
	}

	switch reqEvent.Command {
	case rsTY.CommandGet:
		data, err := svc.getDataRepository(reqEvent)
		if err != nil {
			resEvent.Error = err.Error()
		}
		resEvent.SetData(data)

	default:
		return errors.New("unknown command")
	}
	return svc.postResponse(reqEvent.ReplyTopic, resEvent)
}

func (svc *ResourceService) getDataRepository(request *rsTY.ServiceEvent) (interface{}, error) {
	if request.ID == "" {
		return nil, errors.New("id not supplied")
	}
	cfg, err := svc.api.DataRepository().GetByID(request.ID)
	if err != nil {
		svc.logger.Debug("data repository get failed", zap.String("id", request.ID), zap.Error(err))
		return nil, err
	}
	return cfg, nil
}

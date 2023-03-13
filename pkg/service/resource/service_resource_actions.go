package resource

import (
	"errors"
	"fmt"

	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

func (svc *ResourceService) resourceActionService(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Command == rsTY.CommandSet {
		data, err := svc.getResourceData(reqEvent)
		if err != nil {
			return err
		}
		svc.logger.Debug("resourceActionService", zap.Any("data", data))

		if data.QuickID != "" {
			quickIDWithType := fmt.Sprintf("%s:%s", data.ResourceType, data.QuickID)
			data.QuickID = quickIDWithType
			return svc.actionAPI.ExecuteActionOnResourceByQuickID(data)
		}
		return svc.actionAPI.ExecuteActionOnResourceByLabels(data)
	}
	return fmt.Errorf("unknown command: %s", reqEvent.Command)
}

func (svc *ResourceService) getResourceData(reqEvent *rsTY.ServiceEvent) (*handlerTY.ResourceData, error) {
	if reqEvent.Data == "" {
		return nil, errors.New("data not supplied")
	}
	data := &handlerTY.ResourceData{}
	err := reqEvent.LoadData(data)
	if err != nil {
		svc.logger.Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil, err
	}

	return data, nil
}

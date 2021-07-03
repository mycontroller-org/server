package resource

import (
	"errors"
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/api/action"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/type"
	rsML "github.com/mycontroller-org/server/v2/pkg/model/resource_service"
	"go.uber.org/zap"
)

func resourceActionService(reqEvent *rsML.ServiceEvent) error {
	if reqEvent.Command == rsML.CommandSet {
		data, err := getResourceData(reqEvent)
		if err != nil {
			return err
		}
		zap.L().Debug("resourceActionService", zap.Any("data", data))

		if data.QuickID != "" {
			quickIDWithType := fmt.Sprintf("%s:%s", data.ResourceType, data.QuickID)
			data.QuickID = quickIDWithType
			return action.ExecuteActionOnResourceByQuickID(data)
		}
		return action.ExecuteActionOnResourceByLabels(data)
	}
	return fmt.Errorf("unknown command: %s", reqEvent.Command)
}

func getResourceData(reqEvent *rsML.ServiceEvent) (*handlerType.ResourceData, error) {
	if reqEvent.Data == nil {
		return nil, errors.New("data not supplied")
	}
	data := &handlerType.ResourceData{}
	err := reqEvent.LoadData(data)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil, err
	}

	return data, nil
}

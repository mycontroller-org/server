package resource

import (
	"errors"
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/api/action"
	rsTY "github.com/mycontroller-org/server/v2/pkg/types/resource_service"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/type"
	"go.uber.org/zap"
)

func resourceActionService(reqEvent *rsTY.ServiceEvent) error {
	if reqEvent.Command == rsTY.CommandSet {
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

func getResourceData(reqEvent *rsTY.ServiceEvent) (*handlerTY.ResourceData, error) {
	if reqEvent.Data == nil {
		return nil, errors.New("data not supplied")
	}
	data := &handlerTY.ResourceData{}
	err := reqEvent.LoadData(data)
	if err != nil {
		zap.L().Error("error on data conversion", zap.Any("data", reqEvent.Data), zap.Error(err))
		return nil, err
	}

	return data, nil
}

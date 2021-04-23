package resource

import (
	"errors"
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/api/action"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"go.uber.org/zap"
)

func resourceActionService(reqEvent *rsModel.ServiceEvent) error {
	if reqEvent.Command == rsModel.CommandSet {
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

func getResourceData(reqEvent *rsModel.ServiceEvent) (*handlerML.ResourceData, error) {
	if reqEvent.Data == nil {
		return nil, errors.New("data not supplied")
	}
	data, ok := reqEvent.GetData().(handlerML.ResourceData)
	if !ok {
		return nil, fmt.Errorf("error on data conversion, receivedType: %T", reqEvent.GetData())
	}

	return &data, nil
}

package resource

import (
	"errors"
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/api/action"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/notify_handler"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"go.uber.org/zap"
)

func resourceActionService(reqEvent *rsModel.Event) error {
	if reqEvent.Command == rsModel.CommandSet {
		data, err := getResourceData(reqEvent)
		if err != nil {
			return err
		}
		zap.L().Debug("resourceActionService", zap.Any("data", data))

		if data.QuickID != "" {
			quickIDWithType := fmt.Sprintf("%s:%s", data.ResourceType, data.QuickID)
			return action.ExecuteActionOnResourceByQuickID(quickIDWithType, data.Payload)
		}
		return action.ExecuteActionOnResourceByLabels(data.ResourceType, data.Labels, data.Payload)
	}
	return fmt.Errorf("unknown command: %s", reqEvent.Command)
}

func getResourceData(reqEvent *rsModel.Event) (*handlerML.ResourceData, error) {
	if reqEvent.Data == nil {
		return nil, errors.New("data not supplied")
	}
	var data handlerML.ResourceData
	err := reqEvent.ToStruct(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

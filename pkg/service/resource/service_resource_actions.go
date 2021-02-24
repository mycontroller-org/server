package resource

import (
	"errors"
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/api/action"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"go.uber.org/zap"
)

func resourceActionService(reqEvent *rsModel.Event) error {
	if reqEvent.Command == rsModel.CommandSet {
		data, err := getResourceSelectorData(reqEvent)
		if err != nil {
			return err
		}
		zap.L().Debug("resourceActionService", zap.Any("data", data))

		if data.QuickID != "" {
			return action.ExecuteActionOnResourceByQuickID(data.QuickID, data.Payload)
		}
		return action.ExecuteActionOnResourceByLabels(data.ResourceType, data.Labels, data.Payload)
	}
	return fmt.Errorf("Unknown command: %s", reqEvent.Command)
}

func getResourceSelectorData(reqEvent *rsModel.Event) (*rsModel.ResourceSelector, error) {
	if reqEvent.Data == nil {
		return nil, errors.New("data not supplied")
	}
	var data rsModel.ResourceSelector
	err := reqEvent.ToStruct(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

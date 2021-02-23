package resource

import (
	"errors"
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/api/action"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
	"go.uber.org/zap"
)

func resourceLabelsService(reqEvent *rsModel.Event) error {
	if reqEvent.Command == rsModel.CommandSet {
		data, err := getResourceLabelsData(reqEvent)
		if err != nil {
			return err
		}
		zap.L().Info("resourceLabelsService", zap.Any("data", data))
		return action.ExecuteActionOnResourceByLabels(data.ResourceType, data.Labels, data.Payload)
	}

	return fmt.Errorf("Unknown command: %s", reqEvent.Command)
}

func getResourceLabelsData(reqEvent *rsModel.Event) (*rsModel.ResourceLabels, error) {
	if reqEvent.Data == nil {
		return nil, errors.New("data not supplied")
	}
	var data rsModel.ResourceLabels
	err := reqEvent.ToStruct(&data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

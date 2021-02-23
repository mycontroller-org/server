package resource

import (
	"errors"
	"fmt"

	"github.com/mycontroller-org/backend/v2/pkg/api/action"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	rsModel "github.com/mycontroller-org/backend/v2/pkg/model/resource_service"
)

func resourceQuickIDService(reqEvent *rsModel.Event) error {
	if reqEvent.Command == rsModel.CommandSet {
		quickID, payload, err := getQuickIDData(reqEvent)
		if err != nil {
			return err
		}
		return action.ExecuteActionOnResourceByQuickID(quickID, payload)
	}

	return fmt.Errorf("Unknown command: %s", reqEvent.Command)
}

func getQuickIDData(reqEvent *rsModel.Event) (string, string, error) {
	if reqEvent.Data == nil {
		return "", "", errors.New("data not supplied")
	}
	var data map[string]string
	err := reqEvent.ToStruct(&data)
	if err != nil {
		return "", "", err
	}
	id, ok := data[model.KeyID]
	if !ok {
		return "", "", errors.New("quick id not supplied")
	}

	payload, ok := data[model.KeyPayload]
	if !ok {
		return "", "", errors.New("payload not supplied")
	}
	return id, payload, nil
}

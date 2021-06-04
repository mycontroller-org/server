package googleassistant

import (
	actionAPI "github.com/mycontroller-org/backend/v2/pkg/api/action"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	"github.com/mycontroller-org/backend/v2/pkg/utils/convertor"
	gaML "github.com/mycontroller-org/backend/v2/plugin/bot/google_assistant/model"
	"go.uber.org/zap"
)

func runExecuteRequest(request gaML.ExecuteRequest) *gaML.ExecuteResponse {
	zap.L().Info("received a execute request", zap.Any("request", request))

	response := gaML.ExecuteResponse{RequestID: request.RequestID}
	responseCommands := make([]gaML.ExecuteResponseCommand, 0)
	for _, input := range request.Inputs {
		responseCmd := executePayload(input.Payload)
		responseCommands = append(responseCommands, responseCmd...)
	}

	response.Payload.Commands = responseCommands
	return &response
}

func executePayload(payload gaML.ExecuteRequestPayload) []gaML.ExecuteResponseCommand {
	responseCommands := make([]gaML.ExecuteResponseCommand, 0)
	for _, command := range payload.Commands {
		for _, device := range command.Devices {
			responseCmd := executeCommand(device, command.Execution)
			responseCommands = append(responseCommands, responseCmd...)
		}
	}
	return responseCommands
}

func executeCommand(device gaML.ExecuteRequestDevice, executions []gaML.ExecuteRequestExecution) []gaML.ExecuteResponseCommand {
	responseCommands := make([]gaML.ExecuteResponseCommand, 0)

	for _, execution := range executions {
		var value interface{}
		for key, val := range execution.Params {
			if key == "on" || key == "off" {
				value = val
				break
			}
		}

		if value != nil {
			data := handlerML.ResourceData{
				QuickID: device.ID,
				Payload: convertor.ToString(value),
			}

			statusString := gaML.ExecutionStatusSuccess
			// TODO: include error code
			err := actionAPI.ExecuteActionOnResourceByQuickID(&data)
			if err != nil {
				zap.L().Error("error on executing", zap.Error(err))
				statusString = gaML.ExecutionStatusError
			}
			params := execution.Params
			params["online"] = true
			responseCmd := gaML.ExecuteResponseCommand{
				IDs:    []string{device.ID},
				Status: statusString,
				States: gaML.ExecuteResponseState{Online: true, Others: params},
			}
			responseCommands = append(responseCommands, responseCmd)
		}
	}
	return responseCommands
}

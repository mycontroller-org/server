package googleassistant

import (
	actionAPI "github.com/mycontroller-org/server/v2/pkg/api/action"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	gaTY "github.com/mycontroller-org/server/v2/plugin/bot/google_assistant/types"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

func runExecuteRequest(request gaTY.ExecuteRequest) *gaTY.ExecuteResponse {
	zap.L().Info("received a execute request", zap.Any("request", request))

	response := gaTY.ExecuteResponse{RequestID: request.RequestID}
	responseCommands := make([]gaTY.ExecuteResponseCommand, 0)
	for _, input := range request.Inputs {
		responseCmd := executePayload(input.Payload)
		responseCommands = append(responseCommands, responseCmd...)
	}

	response.Payload.Commands = responseCommands
	return &response
}

func executePayload(payload gaTY.ExecuteRequestPayload) []gaTY.ExecuteResponseCommand {
	responseCommands := make([]gaTY.ExecuteResponseCommand, 0)
	for _, command := range payload.Commands {
		for _, device := range command.Devices {
			responseCmd := executeCommand(device, command.Execution)
			responseCommands = append(responseCommands, responseCmd...)
		}
	}
	return responseCommands
}

func executeCommand(device gaTY.ExecuteRequestDevice, executions []gaTY.ExecuteRequestExecution) []gaTY.ExecuteResponseCommand {
	responseCommands := make([]gaTY.ExecuteResponseCommand, 0)

	for _, execution := range executions {
		var value interface{}
		for key, val := range execution.Params {
			if key == "on" || key == "off" {
				value = val
				break
			}
		}

		if value != nil {
			data := handlerType.ResourceData{
				QuickID: device.ID,
				Payload: convertor.ToString(value),
			}

			statusString := gaTY.ExecutionStatusSuccess
			// TODO: include error code
			err := actionAPI.ExecuteActionOnResourceByQuickID(&data)
			if err != nil {
				zap.L().Error("error on executing", zap.Error(err))
				statusString = gaTY.ExecutionStatusError
			}
			params := execution.Params
			params["online"] = true
			responseCmd := gaTY.ExecuteResponseCommand{
				IDs:    []string{device.ID},
				Status: statusString,
				States: gaTY.ExecuteResponseState{Online: true, Others: params},
			}
			responseCommands = append(responseCommands, responseCmd)
		}
	}
	return responseCommands
}

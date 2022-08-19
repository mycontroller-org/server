package google_assistant

import (
	"fmt"

	actionAPI "github.com/mycontroller-org/server/v2/pkg/api/action"
	vdAPI "github.com/mycontroller-org/server/v2/pkg/api/virtual_device"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	converterUtil "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/types"
	gaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/google/types"
	"go.uber.org/zap"
)

func runExecuteRequest(request gaTY.ExecuteRequest) *gaTY.ExecuteResponse {
	// zap.L().Info("received a execute request", zap.Any("request", request))

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

	vDevice, err := vdAPI.GetByID(device.ID)
	if err != nil {
		return nil
	}

	for _, execution := range executions {
		for key, val := range execution.Params {
			trait, found := gaTY.CommandParamsMap[key]
			if !found {
				zap.L().Warn("trait not found on the command params map", zap.String("virtualDeviceId", vDevice.ID), zap.String("virtualDeviceName", vDevice.Name), zap.String("paramsKey", key))
				continue
			}
			resource, found := vDevice.Traits[trait]
			if !found {
				zap.L().Warn("trait not found on the traits map", zap.String("virtualDeviceId", vDevice.ID), zap.String("virtualDeviceName", vDevice.Name), zap.String("trait", trait))
				continue
			}
			if resource.Type == vdTY.ResourceByQuickID {
				err = actionAPI.ExecuteActionOnResourceByQuickID(&handlerType.ResourceData{
					ResourceType: resource.ResourceType,
					QuickID:      fmt.Sprintf("%s:%s", resource.ResourceType, resource.QuickID),
					Payload:      converterUtil.ToString(val),
					PreDelay:     "0s",
				})
				statusString := gaTY.ExecutionStatusSuccess
				params := execution.Params
				// TODO: include error code
				if err != nil {
					zap.L().Error("error on executing", zap.Error(err))
					statusString = gaTY.ExecutionStatusError
				}
				params[key] = val
				params["online"] = true
				responseCmd := gaTY.ExecuteResponseCommand{
					IDs:    []string{device.ID},
					Status: statusString,
					States: gaTY.ExecuteResponseState{Online: true, Others: params},
				}
				responseCommands = append(responseCommands, responseCmd)
			} else {
				zap.L().Warn("trait not defined with quickId", zap.String("virtualDeviceId", vDevice.ID), zap.String("virtualDeviceName", vDevice.Name), zap.String("trait", trait))
			}
		}

	}
	return responseCommands
}

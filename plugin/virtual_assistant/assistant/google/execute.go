package google_assistant

import (
	"fmt"

	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	gaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/google/types"
	"go.uber.org/zap"
)

func (a *Assistant) runExecuteRequest(request gaTY.ExecuteRequest) *gaTY.ExecuteResponse {
	// a.logger.Info("received a execute request", zap.Any("request", request))

	response := gaTY.ExecuteResponse{RequestID: request.RequestID}
	responseCommands := make([]gaTY.ExecuteResponseCommand, 0)
	for _, input := range request.Inputs {
		responseCmd := a.executePayload(input.Payload)
		responseCommands = append(responseCommands, responseCmd...)
	}

	response.Payload.Commands = responseCommands
	return &response
}

func (a *Assistant) executePayload(payload gaTY.ExecuteRequestPayload) []gaTY.ExecuteResponseCommand {
	responseCommands := make([]gaTY.ExecuteResponseCommand, 0)
	for _, command := range payload.Commands {
		for _, device := range command.Devices {
			responseCmd := a.executeCommand(device, command.Execution)
			responseCommands = append(responseCommands, responseCmd...)
		}
	}
	return responseCommands
}

func (a *Assistant) executeCommand(device gaTY.ExecuteRequestDevice, executions []gaTY.ExecuteRequestExecution) []gaTY.ExecuteResponseCommand {
	responseCommands := make([]gaTY.ExecuteResponseCommand, 0)

	vDevice, err := a.deviceAPI.GetByID(device.ID)
	if err != nil {
		return nil
	}

	for _, execution := range executions {
		for key, val := range execution.Params {
			// if it is in ignore list, keep continue
			if utils.ContainsString(gaTY.IgnoreParamsList, key) {
				continue
			}

			trait, found := gaTY.CommandParamsMap[key]
			if !found {
				a.logger.Warn("trait not found on the command params map", zap.String("virtualDeviceId", vDevice.ID), zap.String("virtualDeviceName", vDevice.Name), zap.String("paramsKey", key))
				continue
			}
			resource, found := vDevice.Traits[trait]
			if !found {
				a.logger.Warn("trait not found on the traits map", zap.String("virtualDeviceId", vDevice.ID), zap.String("virtualDeviceName", vDevice.Name), zap.String("trait", trait))
				continue
			}
			if resource.Type == vdTY.ResourceByQuickID {
				// post data to the actual resource
				quickId := fmt.Sprintf("%s:%s", resource.ResourceType, resource.QuickID)
				err = a.deviceAPI.PostActionOnResourceByQuickID(resource.ResourceType, quickId, val)

				statusString := gaTY.ExecutionStatusSuccess
				params := execution.Params
				// TODO: include error code
				if err != nil {
					a.logger.Error("error on executing", zap.Error(err))
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
				a.logger.Warn("trait not defined with quickId", zap.String("virtualDeviceId", vDevice.ID), zap.String("virtualDeviceName", vDevice.Name), zap.String("trait", trait))
			}
		}

	}
	return responseCommands
}

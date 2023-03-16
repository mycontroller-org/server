package google_assistant

import (
	"github.com/mycontroller-org/server/v2/pkg/types"
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	convertorUtil "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/assistant/google/types"
	"go.uber.org/zap"
)

func (a *Assistant) runQueryRequest(request gaTY.QueryRequest) *gaTY.QueryResponse {
	// a.logger.Info("received a query request", zap.Any("request", request))
	response := gaTY.QueryResponse{
		RequestID: request.RequestID,
		Payload: gaTY.QueryResponsePayload{
			Devices: make(map[string]gaTY.QueryResponseDevice),
		},
	}

	// get device details
	if len(request.Inputs) > 0 {
		input := request.Inputs[0] // for now processing only the 0th index. update the code if there is a chance to receive more than one index
		devicesResp, err := a.queryDevices(input.Payload.Devices)
		if err != nil {
			a.logger.Warn("unable to get devices state", zap.Error(err))
			response.Payload.ErrorCode = err.Error()
		}
		response.Payload.Devices = devicesResp
	}

	return &response
}

func (a *Assistant) queryDevices(devices []gaTY.QueryRequestDevice) (map[string]gaTY.QueryResponseDevice, error) {
	IDs := []string{}
	for _, device := range devices {
		IDs = append(IDs, device.ID)
	}

	devicesResponse := make(map[string]gaTY.QueryResponseDevice)
	if len(IDs) == 0 {
		return devicesResponse, nil
	}

	// get devices
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	vDevices, err := a.deviceAPI.ListDevices(filters, int64(len(IDs)), 0, a.cfg.DeviceFilter)
	if err != nil {
		return nil, err
	}

	// update resource state
	err = a.deviceAPI.UpdateDeviceState(vDevices)
	if err != nil {
		return nil, err
	}

	for _, vDevice := range vDevices {
		response, err := a.queryDeviceState(vDevice)
		if err != nil {
			return nil, err
		}
		devicesResponse[vDevice.ID] = *response
	}

	return devicesResponse, nil
}

func (a *Assistant) queryDeviceState(vDevice vdTY.VirtualDevice) (*gaTY.QueryResponseDevice, error) {
	params := make(map[string]interface{})
	for trait, resource := range vDevice.Traits {
		_params, err := a.getResourceParams(trait, resource)
		if err != nil {
			return nil, err
		}
		utils.JoinMap(params, _params)

	}

	response := gaTY.QueryResponseDevice{
		Online: true,
		Status: "SUCCESS",
		// ErrorCode: ,
		Others: params,
	}
	return &response, nil
}

func (a *Assistant) getResourceParams(trait string, resource vdTY.Resource) (map[string]interface{}, error) {
	// a.logger.Info("requested trait", zap.String("trait", trait))
	params := make(map[string]interface{})
	switch trait {
	case vdTY.DeviceTraitOnOff: // https://developers.google.com/assistant/smarthome/traits/onoff#device-states
		params["on"] = convertorUtil.ToBool(resource.Value)

	case vdTY.DeviceTraitBrightness: // https://developers.google.com/assistant/smarthome/traits/brightness#device-states
		params["brightness"] = convertorUtil.ToInteger(resource.Value)

		// case vdTY.DeviceTraitColorSetting: // https://developers.google.com/assistant/smarthome/traits/colorsetting#device-states
		// 	params["color"] = ""

	default:
		a.logger.Info("support not implemented for this trait", zap.String("trait", trait))
	}
	return params, nil
}

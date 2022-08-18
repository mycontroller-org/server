package google_assistant

import (
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/version"
	botAPI "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/api"
	gaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/google_assistant/types"
	"go.uber.org/zap"
)

const (
	AgentUserId = "1234.12345678"
)

func runSyncRequest(request gaTY.Request) *gaTY.SyncResponse {
	zap.L().Info("received a sync request", zap.Any("request", request))

	devices := make([]gaTY.SyncResponseDevice, 0)
	response := gaTY.SyncResponse{
		RequestID: request.RequestID,
		Payload: gaTY.SyncResponsePayload{
			AgentUserId: AgentUserId,
			Devices:     devices,
		},
	}

	// get virtual devices
	vDevices, err := botAPI.ListDevices(nil, 500, 0) // limits to 500 devices

	if err != nil {
		response.Payload.ErrorCode = err.Error()
	} else {
		for _, vDevice := range vDevices {
			deviceType, found := gaTY.DeviceMap[vDevice.DeviceType]
			if !found {
				zap.L().Info("device type not found in the defined map", zap.String("virtualDeviceId", vDevice.ID), zap.String("virtualDeviceName", vDevice.Name), zap.String("deviceType", vDevice.DeviceType))
				continue
			}
			traits := make([]string, 0)
			for _trait := range vDevice.Traits {
				if trait, found := gaTY.TraitMap[_trait]; found {
					traits = append(traits, trait)
				} else {
					zap.L().Info("trait not found in the defined map", zap.String("virtualDeviceId", vDevice.ID), zap.String("virtualDeviceName", vDevice.Name), zap.String("trait", _trait))
				}
			}
			vDevice.Labels = vDevice.Labels.Init()
			ver := version.Get()
			device := gaTY.SyncResponseDevice{
				ID:                           vDevice.ID,
				Type:                         deviceType,
				Traits:                       traits,
				Name:                         gaTY.NameData{Name: vDevice.Name},
				Attributes:                   nil,
				WillReportState:              false,
				DeviceInfo:                   gaTY.DeviceInfo{Manufacturer: "MyController", SwVersion: ver.Version},
				NotificationSupportedByAgent: false,
				RoomHint:                     vDevice.Labels.Get(types.LabelRoom),
				OtherDeviceIds:               nil,
				CustomData:                   nil,
			}
			devices = append(devices, device)
		}
	}

	zap.L().Info("received a sync request", zap.Any("response", response))

	response.Payload.Devices = devices
	return &response
}

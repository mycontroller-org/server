package googleassistant

import (
	gaTY "github.com/mycontroller-org/server/v2/plugin/bot/google_assistant/types"
	"go.uber.org/zap"
)

func runSyncRequest(request gaTY.Request) *gaTY.SyncResponse {
	zap.L().Info("received a sync request", zap.Any("request", request))

	device := gaTY.SyncResponseDevice{
		ID:   "field:gateway.node.source.field",
		Type: "action.devices.types.LIGHT",
		Traits: []string{
			"action.devices.traits.OnOff",
		},
		Name:                         gaTY.NameData{Name: "night lamp"},
		Attributes:                   nil,
		WillReportState:              true,
		DeviceInfo:                   gaTY.DeviceInfo{},
		NotificationSupportedByAgent: false,
		RoomHint:                     "first floor",
		OtherDeviceIds:               nil,
		CustomData:                   nil,
	}

	response := gaTY.SyncResponse{
		RequestID: request.RequestID,
		Payload: gaTY.SyncResponsePayload{
			AgentUserId: "1234.12345678",
			Devices:     []gaTY.SyncResponseDevice{device},
		},
	}

	zap.L().Info("received a sync request", zap.Any("response", response))

	return &response
}

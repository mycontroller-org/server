package googleassistant

import (
	gaML "github.com/mycontroller-org/server/v2/plugin/bot/google_assistant/model"
	"go.uber.org/zap"
)

func runSyncRequest(request gaML.Request) *gaML.SyncResponse {
	zap.L().Info("received a sync request", zap.Any("request", request))

	device := gaML.SyncResponseDevice{
		ID:   "field:gateway.node.source.field",
		Type: "action.devices.types.LIGHT",
		Traits: []string{
			"action.devices.traits.OnOff",
		},
		Name:                         gaML.NameData{Name: "night lamp"},
		Attributes:                   nil,
		WillReportState:              true,
		DeviceInfo:                   gaML.DeviceInfo{},
		NotificationSupportedByAgent: false,
		RoomHint:                     "first floor",
		OtherDeviceIds:               nil,
		CustomData:                   nil,
	}

	response := gaML.SyncResponse{
		RequestID: request.RequestID,
		Payload: gaML.SyncResponsePayload{
			AgentUserId: "1234.12345678",
			Devices:     []gaML.SyncResponseDevice{device},
		},
	}

	zap.L().Info("received a sync request", zap.Any("response", response))

	return &response
}

package model

// The request same as default Request struct
// so no need to redefine here

// SyncResponse struct
// https://developers.google.com/assistant/smarthome/reference/intent/sync#response
type SyncResponse struct {
	RequestID string              `json:"requestId"` // required
	Payload   SyncResponsePayload `json:"payload"`   // required
}

type SyncResponsePayload struct {
	AgentUserId string               `json:"agentUserId"`           // required
	ErrorCode   string               `json:"errorCode,omitempty"`   // optional
	DebugString string               `json:"debugString,omitempty"` // optional
	Devices     []SyncResponseDevice `json:"devices"`               // required
}

// device details

type SyncResponseDevice struct {
	ID                           string                 `json:"id"`                           // required
	Type                         string                 `json:"type"`                         // required
	Traits                       []string               `json:"traits"`                       // required
	Name                         NameData               `json:"name"`                         // required
	WillReportState              bool                   `json:"willReportState"`              // required
	NotificationSupportedByAgent bool                   `json:"notificationSupportedByAgent"` // default false
	RoomHint                     string                 `json:"roomHint,omitempty"`           // optional
	DeviceInfo                   DeviceInfo             `json:"deviceInfo,omitempty"`         // optional
	Attributes                   map[string]interface{} `json:"attributes,omitempty"`         // optional
	CustomData                   map[string]interface{} `json:"customData,omitempty"`         // optional, maximum of 512 bytes per device
	OtherDeviceIds               []OtherDeviceId        `json:"otherDeviceIds,omitempty"`     // optional
}

type NameData struct {
	Name         string   `json:"name"`                   // required
	DefaultNames []string `json:"defaultNames,omitempty"` // optional
	Nicknames    []string `json:"nicknames,omitempty"`    // optional
}

type DeviceInfo struct {
	Manufacturer string `json:"manufacturer,omitempty"` // optional
	Model        string `json:"model,omitempty"`        // optional
	HwVersion    string `json:"hwVersion,omitempty"`    // optional
	SwVersion    string `json:"swVersion,omitempty"`    // optional
}

type OtherDeviceId struct {
	AgentId       string `json:"agentId,omitempty"` // optional, The agent's ID. Generally, this is the project ID in the Actions console.
	OtherDeviceId string `json:"otherDeviceId"`     // required, Device ID defined by the agent. The device ID must be unique.
}

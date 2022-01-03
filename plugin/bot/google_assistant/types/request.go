package types

// Intent types
const (
	IntentSync       = "action.devices.SYNC"
	IntentQuery      = "action.devices.QUERY"
	IntentExecute    = "action.devices.EXECUTE"
	IntentDisconnect = "action.devices.DISCONNECT"
)

// Request struct
// https://developers.google.com/assistant/smarthome/reference/intent/sync#request
type Request struct {
	RequestID string  `json:"requestId"`
	Inputs    []Input `json:"inputs"`
}

// Input struct
type Input struct {
	Intent string `json:"intent"`
}

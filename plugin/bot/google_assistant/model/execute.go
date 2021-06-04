package model

// Execution status
const (
	ExecutionStatusSuccess    = "SUCCESS"
	ExecutionStatusPending    = "PENDING"
	ExecutionStatusOffline    = "OFFLINE"
	ExecutionStatusExceptions = "EXCEPTIONS"
	ExecutionStatusError      = "ERROR"
)

// ExecuteRequest struct
// https://developers.google.com/assistant/smarthome/reference/intent/execute#request
type ExecuteRequest struct {
	RequestID string                `json:"requestId"` // required
	Inputs    []ExecuteRequestInput `json:"inputs"`    // required
}

type ExecuteRequestInput struct {
	Intent  string                `json:"intent"`  // required
	Payload ExecuteRequestPayload `json:"payload"` // required
}

type ExecuteRequestPayload struct {
	Commands []ExecuteRequestCommand `json:"commands"` // required
}

type ExecuteRequestCommand struct {
	Devices   []ExecuteRequestDevice    `json:"devices"`   // required
	Execution []ExecuteRequestExecution `json:"execution"` // required
}

type ExecuteRequestDevice struct {
	ID         string                 `json:"id"`                   // required
	CustomData map[string]interface{} `json:"customData,omitempty"` // optional
}

type ExecuteRequestExecution struct {
	Command string                 `json:"command"`          // required
	Params  map[string]interface{} `json:"params,omitempty"` // optional
}

// ExecuteResponse struct
// https://developers.google.com/assistant/smarthome/reference/intent/execute#response
type ExecuteResponse struct {
	RequestID string                 `json:"requestId"` // required
	Payload   ExecuteResponsePayload `json:"payload"`   // required
}

type ExecuteResponsePayload struct {
	ErrorCode   string                   `json:"errorCode,omitempty"`   // optional
	DebugString string                   `json:"debugString,omitempty"` // optional
	Commands    []ExecuteResponseCommand `json:"commands"`              // required
}

type ExecuteResponseCommand struct {
	IDs       []string             `json:"ids"`              // required
	Status    string               `json:"status"`           // required
	States    ExecuteResponseState `json:"states,omitempty"` // optional
	ErrorCode string               `json:"errorCode"`        // optional
}

type ExecuteResponseState struct {
	Online bool                   `json:"online,omitempty"` // optional
	Others map[string]interface{} `json:",inline"`          // optional
}

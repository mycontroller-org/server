package types

import (
	"encoding/json"

	"go.uber.org/zap"
)

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
	Others map[string]interface{} `json:"-"`                // optional, `json:",inline"` does not work, https://github.com/golang/go/issues/6213
}

// reference taken from: https://stackoverflow.com/questions/49901287/embed-mapstringstring-in-go-json-marshaling-without-extra-json-property-inlin
func (ers ExecuteResponseState) MarshalJSON() ([]byte, error) {
	// Turn qrd into a map
	type ExecuteResponseState_ ExecuteResponseState // prevent recursion
	b, err := json.Marshal(ExecuteResponseState_(ers))
	if err != nil {
		zap.L().Error("error on marshal ExecuteResponseState", zap.Error(err))
		return nil, err
	}

	var m map[string]json.RawMessage
	err = json.Unmarshal(b, &m)
	if err != nil {
		zap.L().Error("error on Unmarshal ExecuteResponseState", zap.Error(err))
		return nil, err
	}

	// Add tags to the map, possibly overriding struct fields
	for k, v := range ers.Others {
		// if overriding struct fields is not acceptable:
		if _, ok := m[k]; ok {
			continue
		}
		b, err = json.Marshal(v)
		if err != nil {
			zap.L().Error("error on marshal ExecuteResponseState Others field", zap.String("key", k), zap.Error(err))
			return nil, err
		}
		m[k] = b
	}

	return json.Marshal(m)
}

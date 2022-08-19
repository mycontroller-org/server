package types

import (
	"encoding/json"

	"go.uber.org/zap"
)

const (
	StatusSuccess    = "SUCCESS"
	StatusOffline    = "OFFLINE"
	StatusExceptions = "EXCEPTIONS"
	StatusError      = "ERROR"
)

// QueryRequest struct
// https://developers.google.com/assistant/smarthome/reference/intent/query#request
type QueryRequest struct { // required
	RequestID string              `json:"requestId"` // required
	Inputs    []QueryRequestInput `json:"inputs"`    // required
}

// QueryRequestInput struct
type QueryRequestInput struct {
	Intent  string              `json:"intent"`  // required
	Payload QueryRequestPayload `json:"payload"` // required
}

type QueryRequestPayload struct {
	Devices []QueryRequestDevice `json:"devices"` // required
}

type QueryRequestDevice struct {
	ID         string                 `json:"id"`                   // required
	CustomData map[string]interface{} `json:"customData,omitempty"` // optional
}

// QueryResponse
// https://developers.google.com/assistant/smarthome/reference/intent/query#response
type QueryResponse struct {
	RequestID string               `json:"requestId"` // required
	Payload   QueryResponsePayload `json:"payload"`   // required
}

type QueryResponsePayload struct {
	ErrorCode   string                         `json:"errorCode,omitempty"`   // optional
	DebugString string                         `json:"debugString,omitempty"` // optional
	Devices     map[string]QueryResponseDevice `json:"devices"`               // required
}

// QueryResponseDevice struct
type QueryResponseDevice struct {
	Online    bool                   `json:"online"`              // required
	Status    string                 `json:"status"`              // required
	ErrorCode string                 `json:"errorCode,omitempty"` // optional
	Others    map[string]interface{} `json:"-"`                   // optional, `json:",inline"` does not work, https://github.com/golang/go/issues/6213
}

// Note: QueryResponseDevice supports extra fields. Add those fields in others
// Example:
// {
// 		"on": true,
// 		"online": true,
// 		"status": "SUCCESS",
// 		"brightness": 80,
// 		"color": {
// 			"name": "cerulean",
// 			"spectrumRGB": 31655
// 		}
// }

// reference taken from: https://stackoverflow.com/questions/49901287/embed-mapstringstring-in-go-json-marshaling-without-extra-json-property-inlin
func (qrd QueryResponseDevice) MarshalJSON() ([]byte, error) {
	// Turn qrd into a map
	type QueryResponseDevice_ QueryResponseDevice // prevent recursion
	b, err := json.Marshal(QueryResponseDevice_(qrd))
	if err != nil {
		zap.L().Error("error on marshal QueryResponseDevice", zap.Error(err))
		return nil, err
	}

	var m map[string]json.RawMessage
	err = json.Unmarshal(b, &m)
	if err != nil {
		zap.L().Error("error on Unmarshal QueryResponseDevice", zap.Error(err))
		return nil, err
	}

	// Add tags to the map, possibly overriding struct fields
	for k, v := range qrd.Others {
		// if overriding struct fields is not acceptable:
		if _, ok := m[k]; ok {
			continue
		}
		b, err = json.Marshal(v)
		if err != nil {
			zap.L().Error("error on marshal QueryResponseDevice Others field", zap.String("key", k), zap.Error(err))
			return nil, err
		}
		m[k] = b
	}

	return json.Marshal(m)
}

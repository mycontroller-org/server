package types

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
	Others    map[string]interface{} `json:",inline"`             // optional
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

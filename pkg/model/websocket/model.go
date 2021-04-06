package websocket

// Response type
const (
	ResponseTypeResource = "resource"
)

// Resource struct
type Resource struct {
	Type     string      `json:"type"`
	ID       string      `json:"id"`
	QuickID  string      `json:"quickId"`
	Resource interface{} `json:"resource"`
}

// Response of a websocket
type Response struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Request for websocket
type Request struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// SubscribeRequest details
type SubscribeRequest struct {
	Resources []Resource `json:"resources"`
}

// unsubscribeRequest details
type UnsubscribeRequest struct {
	Resources []Resource `json:"resources"`
}

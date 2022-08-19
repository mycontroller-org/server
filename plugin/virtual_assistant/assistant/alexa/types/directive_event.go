package types

type DirectiveOrEvent struct {
	Header   Header                 `json:"header"`
	Payload  map[string]interface{} `json:"payload"`
	Endpoint *DirectiveEndpoint     `json:"endpoint,omitempty"`
}

type DirectiveEndpoint struct {
	Scope      Scope             `json:"scope"`
	EndpointID string            `json:"endpointId"`
	Cookie     map[string]string `json:"cookie"`
}

type Scope struct {
	Type      string `json:"type"`
	Token     string `json:"token"`
	Partition string `json:"partition,omitempty"`
	UserID    string `json:"userId,omitempty"`
}

type Header struct {
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	MessageID      string `json:"messageId"`
	PayloadVersion string `json:"payloadVersion"`
}

package types

type Response struct {
	Event   DirectiveOrEvent `json:"event"`
	Context *Context         `json:"context,omitempty"`
}

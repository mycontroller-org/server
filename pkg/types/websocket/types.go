package websocket

import (
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/bus/event"
)

// Request types
const (
	RequestTypeSubscribeEvent   = "subscribe_event"
	RequestTypeUnsubscribeEvent = "unsubscribe_event"
)

// Response types
const (
	ResponseTypeEvent = "event"
)

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
	Resources []eventTY.Event `json:"events"`
}

// unsubscribeRequest details
type UnsubscribeRequest struct {
	Resources []eventTY.Event `json:"events"`
}

package websocket

import (
	eventML "github.com/mycontroller-org/server/v2/pkg/model/bus/event"
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
	Resources []eventML.Event `json:"events"`
}

// unsubscribeRequest details
type UnsubscribeRequest struct {
	Resources []eventML.Event `json:"events"`
}

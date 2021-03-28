package mcbus

import "fmt"

// topics used across the application
const (
	TopicPostMessageToCore             = "message.to_core"                     // processor listens. posts message in to core component
	TopicPostMessageToProvider         = "message.to_provider"                 // provider listens. append gateway id
	TopicPostRawMessageAcknowledgement = "message.raw_message_acknowledgement" // raw message acknowledge
	TopicPostMessageNotifyHandler      = "message.notify_handler"              // post to notify handler
	TopicServiceResourceServer         = "service.resource_server"             // a server listens on this topic, and serves the request
	TopicServiceGateway                = "service.gateway"                     // gateways listen this topic and perform actions like load, reload, stop, start, etc.,
	TopicServiceNotifyHandler          = "service.notify_handler"              // notify handler listen this topic and perform actions like load, reload, stop, start, etc.,
	TopicServiceTask                   = "service.task"                        // tasks listen this topic and perform actions like load, reload, stop, start, etc.,
	TopicServiceScheduler              = "service.scheduler"                   // scheduler listen this topic and perform actions like add, remove, disable, etc.,
	TopicEventsAll                     = "event.>"                             // all events
	TopicEventGateway                  = "event.gateway"                       // gateway events
	TopicEventNode                     = "event.node"                          // node events
	TopicEventSource                   = "event.source"                        // source events
	TopicEventFieldSet                 = "event.field.set"                     // field set events
	TopicEventFieldRequest             = "event.field.request"                 // field request events
)

const keyTopicPrefix = "topic_prefix"

var topicPrefix = ""

// FormatTopic adds prefix and arguments
func FormatTopic(topic string, arguments ...interface{}) string {
	formated := topic
	if len(arguments) > 0 {
		formated = fmt.Sprintf(topic, arguments...)
	}
	if topicPrefix != "" {
		return fmt.Sprintf("%s.%s", topicPrefix, formated)
	}
	return formated
}

// GetTopicPostMessageToCore posts messages to provider
func GetTopicPostMessageToCore() string {
	return FormatTopic(TopicPostMessageToCore)
}

// GetTopicPostMessageToProvider posts messages to provider
func GetTopicPostMessageToProvider(gatewayID string) string {
	return FormatTopic("%s.%s", TopicPostMessageToProvider, gatewayID)
}

// GetTopicPostRawMessageAcknowledgement posts ack, used in provider (if needed)
func GetTopicPostRawMessageAcknowledgement(gatewayID, msgID string) string {
	return FormatTopic("%s.%s.%s", TopicPostRawMessageAcknowledgement, gatewayID, msgID)
}

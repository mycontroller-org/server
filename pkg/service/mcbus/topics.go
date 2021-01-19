package mcbus

import "fmt"

// topics used across the application
const (
	TopicPostMessageToCore             = "message_to_core"             // processor listens. posts message in to core component
	TopicPostMessageToProvider         = "message_to_provider"         // provider listens. append gateway id
	TopicPostRawMessageAcknowledgement = "raw_message_acknowledgement" // raw message acknowledge
	TopicResourceServer                = "resource_server"             // a server listens on this topic, and serves the request
	TopicServiceGateway                = "service_gateway"             // gateways listen this topic and perform actions like load, reload, stop, start, etc.,
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
		return fmt.Sprintf("%s/%s", topicPrefix, formated)
	}
	return formated
}

// GetTopicPostMessageToCore posts messages to provider
func GetTopicPostMessageToCore() string {
	return FormatTopic(TopicPostMessageToCore)
}

// GetTopicPostMessageToProvider posts messages to provider
func GetTopicPostMessageToProvider(gatewayID string) string {
	return FormatTopic("%s/%s", TopicPostMessageToProvider, gatewayID)
}

// GetTopicPostRawMessageAcknowledgement posts ack, used in provider (if needed)
func GetTopicPostRawMessageAcknowledgement(gatewayID, msgID string) string {
	return FormatTopic("%s/%s/%s", TopicPostRawMessageAcknowledgement, gatewayID, msgID)
}

package mcbus

// global gateway topics
const (
	TopicMessageToGateway               = "message_to_gateway"                 // append gateway id
	TopicSleepingMessageToGateway       = "sleeping_message_to_gateway"        // append gateway id
	TopicGatewayAcknowledgement         = "provider_acknowledgement"           // append with gateway id and message id
	TopicMessageFromGateway             = "message_from_gateway"               //
	TopicMessageToGatewayDelieverStatus = "message_to_gateway_deliever_status" // append with message id
)

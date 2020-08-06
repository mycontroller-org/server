package mcbus

// global gateway topics
const (
	TopicMsg2GW                 = "msg_to_gw"                 // append gateway id
	TopicSleepingMsg2GW         = "sleeping_msg_to_gw"        // append gateway id
	TopicGatewayAcknowledgement = "gw_ack"                    // append with message id
	TopicMsgFromGW              = "msg_from_gw"               //
	TopicMsg2GWDelieverStatus   = "msg_to_gw_deliever_status" // append with message id
)

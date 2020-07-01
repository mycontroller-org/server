package service

// topics used across the application
const (
	TopicGatewayMessageReceive  = "t_gw_msg_rx"
	TopicGatewayMessageTransmit = "t_gw_msg_tx"
)

func registerTopics() {
	BUS.RegisterTopics(TopicGatewayMessageReceive)
}

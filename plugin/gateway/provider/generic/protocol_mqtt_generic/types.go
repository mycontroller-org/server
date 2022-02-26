package mqtt_generic

import (
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	mqtt "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/protocol_mqtt"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
)

const (
	ScriptKeyDataIn   = "dataIn"
	ScriptKeyDataOut  = "dataOut"
	ScriptKeyTopicIn  = "topicIn"
	ScriptKeyTopicOut = "topicOut"
	ScriptKeyConfigIn = "configIn"

	DefaultNode = "default"
)

type MqttProtocol struct {
	GatewayCfg *gwTY.Config
	Protocol   gwPtl.Protocol
	Config     *MqttProtocolConf
	rxMsgFunc  func(rm *msgTY.RawMessage) error
}

type MqttProtocolConf struct {
	mqtt.Config
	Nodes map[string]MqttNode `json:"nodes"`
}

// mqtt node config
type MqttNode struct {
	Topic  string `json:"topic"`
	QoS    int    `json:"qos"`
	Script string `json:"script"`
}

// Clone cones the MqttNode
func (mn *MqttNode) Clone() *MqttNode {
	cloned := &MqttNode{
		Topic:  mn.Topic,
		QoS:    mn.QoS,
		Script: mn.Script,
	}
	return cloned
}

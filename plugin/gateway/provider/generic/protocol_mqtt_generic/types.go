package mqtt_generic

import (
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	mqtt "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/protocol_mqtt"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
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
	logger     *zap.Logger
}

type MqttProtocolConf struct {
	mqtt.Config
	Nodes map[string]MqttNode `json:"nodes" yaml:"nodes"`
}

// mqtt node config
type MqttNode struct {
	Topic  string `json:"topic" yaml:"topic"`
	QoS    int    `json:"qos" yaml:"qos"`
	Script string `json:"script" yaml:"script"`
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

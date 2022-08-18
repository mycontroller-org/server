package mqtt_generic

import (
	"fmt"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	jsUtils "github.com/mycontroller-org/server/v2/pkg/utils/javascript"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	mqtt "github.com/mycontroller-org/server/v2/plugin/gateway/protocol/protocol_mqtt"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	"go.uber.org/zap"
)

// returns a new generic mqtt protocol
func New(gwCfg *gwTY.Config, protocol cmap.CustomMap, rxMsgFunc func(rm *msgTY.RawMessage) error) (*MqttProtocol, error) {
	config := &MqttProtocolConf{}
	err := json.ToStruct(protocol, config)
	if err != nil {
		zap.L().Error("error on converting to protocol config", zap.String("gatewayId", gwCfg.ID), zap.Error(err))
		return nil, err
	}

	mp := &MqttProtocol{
		GatewayCfg: gwCfg,
		Config:     config,
		rxMsgFunc:  rxMsgFunc,
	}

	mqttBaseProtocol, err := mqtt.New(gwCfg, protocol, mp.onMessageReceive)
	if err != nil {
		zap.L().Error("error on getting base mqtt protocol", zap.String("gatewayId", gwCfg.ID), zap.Error(err))
		return nil, err
	}
	mp.Protocol = mqttBaseProtocol

	return mp, nil
}

// posts received messages in to queue
func (mp *MqttProtocol) onMessageReceive(rawMessage *msgTY.RawMessage) error {
	// convert bytes to hex string
	stringData := fmt.Sprintf("%X", rawMessage.Data)
	rawMessage.Data = stringData

	return mp.rxMsgFunc(rawMessage)
}

// posts a message to a topic
func (mp *MqttProtocol) Post(msg *msgTY.Message) error {
	cfgRaw, ok := mp.Config.Nodes[msg.NodeID]
	if !ok {
		defaultCfg, ok := mp.Config.Nodes[DefaultNode]
		if !ok {
			return fmt.Errorf("node not defined, nodeID:%s", msg.NodeID)
		}
		cfgRaw = defaultCfg
	}

	endpoint := &MqttNode{}
	err := json.ToStruct(cfgRaw, endpoint)
	if err != nil {
		zap.L().Error("error on converting to http endpoint config", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
		return err
	}

	endpoint = endpoint.Clone()

	finalMessage := ""
	topic := endpoint.Topic
	if endpoint.Script != "" {
		variables := map[string]interface{}{
			ScriptKeyConfigIn: *endpoint,
			ScriptKeyDataIn:   *msg,
		}

		scriptResponse, err := jsUtils.Execute(endpoint.Script, variables)
		if err != nil {
			zap.L().Error("error on executing script", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
			return err
		}
		mapResponse, err := jsUtils.ToMap(scriptResponse)
		if err != nil {
			zap.L().Error("error on converting to map", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
			return err
		}
		// update topic
		scriptTopic := utils.GetMapValue(mapResponse, ScriptKeyTopicOut, nil)
		if scriptTopic != nil {
			topic = convertor.ToString(scriptTopic)
		}

		// update message
		scriptMessage := utils.GetMapValue(mapResponse, ScriptKeyDataOut, "")
		finalMessage = convertor.ToString(scriptMessage)

	}

	rawMessage := &msgTY.RawMessage{
		Timestamp: time.Now(),
		Others:    cmap.CustomMap{},
		Data:      []byte(finalMessage),
	}

	rawMessage.Others.Set(gwPtl.KeyMqttQoS, endpoint.QoS, nil)
	// update topics
	if topic == "" {
		return fmt.Errorf("empty topic. gatewayId:%s, nodeId:%s", mp.GatewayCfg.ID, msg.NodeID)
	}
	topics := strings.Split(topic, ",")
	rawMessage.Others.Set(gwPtl.KeyMqttTopic, topics, nil)

	// send the message
	return mp.Protocol.Write(rawMessage)
}

// closes the protocol
func (mp *MqttProtocol) Close() error {
	return mp.Protocol.Close()
}

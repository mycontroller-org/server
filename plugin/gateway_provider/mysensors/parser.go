package mysensors

import (
	"errors"
	"fmt"
	"strings"

	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	msg "github.com/mycontroller-org/mycontroller-v2/pkg/model/message"
	"go.uber.org/zap"
)

// Parser implementation
type Parser struct {
	Gateway *ml.GatewayConfig
}

// ToRawMessage converts to gateway specific
func (p *Parser) ToRawMessage(mcMsg *msg.Message) (*msg.RawMessage, error) {
	msMsg := myMessage{
		NodeID:   mcMsg.NodeID,
		SensorID: mcMsg.SensorID,
		Command:  "",
		Ack:      "0",
		Type:     "",
		Payload:  "",
	}
	if p.Gateway.AckConfig.Enabled {
		msMsg.Ack = "1"
	}

	if len(mcMsg.Payload) > 0 {
		rm := &msg.RawMessage{
			Timestamp: mcMsg.Timestamp,
			Others:    map[string]interface{}{},
		}
		// get command
		msMsg.Type = mcMsg.Others[keyCmdType].(string)

		// create rawMessage
		switch p.Gateway.Provider.GatewayType {
		case ml.GatewayTypeSerial, ml.GatewayTypeEthernet:

		case ml.GatewayTypeMQTT:
			rm.Data = []byte(mcMsg.Payload)
			rm.Others[msg.KeyTopic] = msMsg.toRaw(true)
		}
		return rm, nil
	}
	return nil, errors.New("No fields found in the given input")
}

// ToMessage converts to mc specific
func (p *Parser) ToMessage(rm *msg.RawMessage) (*msg.Message, error) {
	//zap.L().Debug("Raw message", zap.Any("rawMessage", rm))

	d := make([]string, 0)
	payload := ""

	// decode message from gateway
	switch p.Gateway.Provider.GatewayType {
	case ml.GatewayTypeMQTT:
		// topic/node-id/child-sensor-id/command/ack/type
		// out_rfm69/11/1/1/0/0
		_d := strings.Split(string(rm.Others[msg.KeyTopic].(string)), "/")
		if len(_d) < 5 {
			zap.L().Error("Invalid message format", zap.Any("message", rm))
			return nil, nil
		}
		d = _d[len(_d)-5:]
		payload = string(rm.Data)
	case ml.GatewayTypeSerial, ml.GatewayTypeEthernet:
		// node-id;child-sensor-id;command;ack;type;payload
		_d := strings.Split(string(rm.Data), ";")
		if len(_d) < 6 {
			zap.L().Error("Invalid message format", zap.String("message", string(rm.Data)))
			return nil, nil
		}
		payload = _d[6]
		d = _d[5:]
	// implement this one
	default:
		return nil, fmt.Errorf("This gateway type not implements. gatewayType: %s", p.Gateway.Provider.GatewayType)
	}

	//zap.L().Debug("message", zap.Any("slice", d))

	ms := myMessage{
		NodeID:   d[0],
		SensorID: d[1],
		Command:  d[2],
		Ack:      d[3],
		Type:     d[4],
		Payload:  payload,
	}

	mcMsg := &msg.Message{
		NodeID:     ms.NodeID,
		SensorID:   ms.SensorID,
		IsAck:      ms.Ack == "1",
		IsReceived: true,
		Timestamp:  rm.Timestamp,
		Payload:    ms.Payload,
		Command:    cmdMapIn[ms.Command],
		Others:     make(map[string]interface{}),
	}

	switch ms.Command {
	case CmdSet, CmdReq:
		mcMsg.Field = cmdSetReqTypeMapIn[ms.Type]
		_tu := metricUnit[mcMsg.Field]
		mcMsg.PayloadType = _tu.Type
		mcMsg.PayloadUnitID = _tu.Unit
		mcMsg.Others[keyCmdType] = ms.Type
	case CmdInternal:
		// check should I have to handle this locally?
		if fn, ok := localHandlerMapIn[ms.Type]; ok {
			nms := fn(p.Gateway, ms)
			if nms != nil {
				// TODO: send this message to node and update last seen
			}
		}
		if _type, ok := cmdInternalTypeMapIn[ms.Type]; ok {
			mcMsg.SubCommand = _type
		} else {
			// ignore to process this message
			return nil, nil
		}
	case CmdPresentation:
		if _type, ok := cmdPresentationTypeMapIn[ms.Type]; ok {
			if _type == "S_ARDUINO_REPEATER_NODE" || _type == "S_ARDUINO_NODE" {
				// this is a node data
				mcMsg.SubCommand = msg.CommandPresentation
				mcMsg.Field = msg.SubCmdLibraryVersion
				if _type == "S_ARDUINO_REPEATER_NODE" {
					mcMsg.Others[keyNodeType] = "Repeater"
				}
			} else {
				mcMsg.Field = msg.KeyName
				mcMsg.Others[keyCmdType] = ms.Type
				mcMsg.Others[KeyCmdTypeString] = _type
			}
		} else {
			return nil, nil
		}
	}
	return mcMsg, nil
}

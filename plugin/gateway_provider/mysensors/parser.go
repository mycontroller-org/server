package mysensors

import (
	"errors"
	"fmt"
	"strings"

	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gateway_protocol"
	"go.uber.org/zap"
)

// ToRawMessage converts to gateway specific
func (p *Provider) ToRawMessage(msg *msgml.Message) (*msgml.RawMessage, error) {
	msMsg := message{
		NodeID:   msg.NodeID,
		SensorID: msg.SensorID,
		Command:  "",
		Ack:      "0",
		Type:     "",
		Payload:  "",
	}

	// init labels and others
	msg.Labels = msg.Labels.Init()
	msg.Others = msg.Others.Init()

	if p.GWConfig.Ack.Enabled {
		msMsg.Ack = "1"
	}

	rawMsg := &msgml.RawMessage{Timestamp: msg.Timestamp}
	rawMsg.Others = rawMsg.Others.Init()

	// get command
	msMsg.Type = msg.Labels.Get(keyType)
	if msMsg.Type == "" {
		return nil, errors.New("command type should not be empty")
	}

	// create rawMessage
	switch p.GWConfig.Provider.ProtocolType {
	case gwpl.TypeSerial, gwpl.TypeEthernet:

	case gwpl.TypeMQTT:
		rawMsg.Data = []byte(msg.Payload)
		rawMsg.Others.Set(gwpl.KeyTopic, msMsg.toRaw(true), nil)
	}
	return rawMsg, nil
}

// ToMessage converts to mc specific
func (p *Provider) ToMessage(rawMsg *msgml.RawMessage) (*msgml.Message, error) {

	rawMsg.Others = rawMsg.Others.Init()

	d := make([]string, 0)
	payload := ""

	// decode message from gateway
	switch p.GWConfig.Provider.ProtocolType {
	case gwpl.TypeMQTT:
		// topic/node-id/child-sensor-id/command/ack/type
		// out_rfm69/11/1/1/0/0
		rData := strings.Split(string(rawMsg.Others[gwpl.KeyTopic].(string)), "/")
		if len(rData) < 5 {
			zap.L().Error("Invalid message format", zap.Any("rawMessage", rawMsg))
			return nil, nil
		}
		d = rData[len(rData)-5:]
		payload = string(rawMsg.Data)
	case gwpl.TypeSerial, gwpl.TypeEthernet:
		// node-id;child-sensor-id;command;ack;type;payload
		_d := strings.Split(string(rawMsg.Data), ";")
		if len(_d) < 6 {
			zap.L().Error("Invalid message format", zap.String("rawMessage", string(rawMsg.Data)))
			return nil, nil
		}
		payload = _d[6]
		d = _d[5:]
	// implement this one
	default:
		return nil, fmt.Errorf("This type not implements. protocol type: %s", p.GWConfig.Provider.ProtocolType)
	}

	// MySensors message
	msMsg := message{
		NodeID:   d[0],
		SensorID: d[1],
		Command:  d[2],
		Ack:      d[3],
		Type:     d[4],
		Payload:  payload,
	}

	// Message
	msg := &msgml.Message{
		NodeID:     msMsg.NodeID,
		SensorID:   msMsg.SensorID,
		IsAck:      msMsg.Ack == "1",
		IsReceived: true,
		Timestamp:  rawMsg.Timestamp,
		Payload:    msMsg.Payload,
		Type:       cmdMapIn[msMsg.Command],
	}

	// init labels and others
	msg.Labels = msg.Labels.Init()
	msg.Others = msg.Others.Init()

	// Remove sensor id, if it is a internal message
	if msg.SensorID == idBroadcast {
		msg.SensorID = ""
	}

	// Remove node id, if it is a broadcast message
	if msg.NodeID == idBroadcast {
		msg.NodeID = ""
	}

	// set labels
	msg.Labels.Set(keyType, msMsg.Type)

	switch {

	case msMsg.SensorID != idBroadcast: // perform sensor related stuff
		// update some stuffs as label
		msg.Labels.Set(keyNodeID, msMsg.NodeID)
		msg.Labels.Set(keySensorID, msMsg.SensorID)

		switch msMsg.Command {
		case CmdPresentation:
			if _type, ok := cmdPresentationTypeMapIn[msMsg.Type]; ok {
				msg.FieldName = fml.FieldName
				msg.Labels.Set(KeyTypeString, _type)
			} else {
				// not supported? should I have to return from here?
			}

		case CmdSet, CmdReq:
			_field, ok := cmdSetReqFieldMapIn[msMsg.Type]
			if ok {
				msg.Labels.Set(KeyTypeString, _field)
			} else {
				_field = "V_CUSTOM"
				zap.L().Warn("This set, req not found. update this. Setting as V_CUSTOM", zap.Any("msMsg", msMsg))
			}

			// get type and unit
			if typeUnit, ok := metricTypeUnit[_field]; ok {
				msg.FieldName = _field
				msg.MetricType = typeUnit.Type
				msg.Unit = typeUnit.Unit
			} else {
				// not supported? should I have to return from here?
			}

		default:
			// not supported? should I have to return from here?
		}

	case msMsg.NodeID != idBroadcast: // perform node related stuff
		// update some stuffs as label
		msg.Labels.Set(keyNodeID, msMsg.NodeID)
		switch msMsg.Command {

		case CmdPresentation:
			msg.Others.Set(fml.FieldLibraryVersion, msg.Payload, nil) // set lib version
			if _type, ok := cmdPresentationTypeMapIn[msMsg.Type]; ok {
				if _type == "S_ARDUINO_REPEATER_NODE" || _type == "S_ARDUINO_NODE" {
					// this is a node data
					msg.FieldName = fml.FieldName
					if _type == "S_ARDUINO_REPEATER_NODE" {
						msg.Labels.Set(keyNodeType, "repeater")
					}
				} else {
					// return?
				}
			} else {
				// return?
			}
		case CmdInternal:
			if _type, ok := cmdInternalTypeMapIn[msMsg.Type]; ok {
				if fieldName, ok := cmdInternalFieldMap[_type]; ok {
					msg.FieldName = fieldName
					msg.Type = msgml.TypeSet
				} else { // this will be implemented properly with functions call
					msg.FieldName = _type
					msg.Type = msgml.TypeInternal
				}
			} else {
				// return?
			}
		}
	default:
		// if none of the above
	}

	return msg, nil
}

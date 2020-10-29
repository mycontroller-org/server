package mysensors

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	mtrml "github.com/mycontroller-org/backend/v2/pkg/model/metric"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	"github.com/mycontroller-org/backend/v2/pkg/util"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gw_protocol"
	"go.uber.org/zap"
)

// ToRawMessage converts to gateway specific
func (p *Provider) ToRawMessage(msg *msgml.Message) (*msgml.RawMessage, error) {
	if len(msg.Payloads) == 0 {
		return nil, errors.New("There is no payload details on the message")
	}

	payload := msg.Payloads[0]

	msMsg := message{
		NodeID:   msg.NodeID,
		SensorID: msg.SensorID,
		Command:  "",
		Ack:      "0",
		Type:     "",
		Payload:  payload.Value,
	}

	// update broadcast id
	if msMsg.NodeID == "" {
		msMsg.NodeID = idBroadcast
	}
	if msMsg.SensorID == "" {
		msMsg.SensorID = idBroadcast
	}

	// init labels and others
	payload.Labels = payload.Labels.Init()
	payload.Others = payload.Others.Init()

	if msg.IsAckEnabled {
		msMsg.Ack = "1"
	}

	rawMsg := &msgml.RawMessage{Timestamp: msg.Timestamp}
	rawMsg.Others = rawMsg.Others.Init()

	// get command
	switch msg.Type {

	case msgml.TypeSet:
		msMsg.Command = cmdSet
		msMsg.Type = payload.Labels.Get(LabelType)
		if msMsg.Type == "" {
			for k, v := range setReqFieldMapForRx {
				if v == payload.Name {
					msMsg.Type = k
					break
				}
			}
		}
		if mt, ok := metricTypeAndUnit[payload.Name]; ok {
			if mt.Type == mtrml.MetricTypeBinary {
				switch strings.ToLower(payload.Value) {
				case "true", "on":
					msMsg.Payload = payloadON
				case "false", "off":
					msMsg.Payload = payloadOFF
				}
			}
		}

	case msgml.TypeRequest:
		msMsg.Command = cmdRequest
		msMsg.Type = payload.Labels.Get(LabelType)

	case msgml.TypeAction:
		msMsg.Command = cmdInternal
		// call functions
		err := handleActions(p.GWConfig, payload.Name, msg, &msMsg)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("This command not implemented: %s", msg.Type)
	}

	if msMsg.Type == "" {
		return nil, errors.New("command type should not be empty")
	}

	// create rawMessage
	switch p.GWConfig.Provider.ProtocolType {
	case gwpl.TypeSerial, gwpl.TypeEthernet:
		rawMsg.Data = []byte(msMsg.toMySensorsRaw(false))

	case gwpl.TypeMQTT:
		rawMsg.Data = []byte(msMsg.Payload)
		rawMsg.Others.Set(gwpl.KeyMqttTopic, []string{msMsg.toMySensorsRaw(true)}, nil)

	default:
		return nil, fmt.Errorf("This protocol not implemented: %s", p.GWConfig.Provider.ProtocolType)
	}
	return rawMsg, nil
}

// ToMessage converts to mc specific
func (p *Provider) ToMessage(rawMsg *msgml.RawMessage) ([]*msgml.Message, error) {

	messages := make([]*msgml.Message, 0)

	d := make([]string, 0)
	payload := ""

	// decode message from gateway
	switch p.GWConfig.Provider.ProtocolType {
	case gwpl.TypeMQTT:
		// topic/node-id/child-sensor-id/command/ack/type
		// out_rfm69/11/1/1/0/0
		rData := strings.Split(string(rawMsg.Others.Get(gwpl.KeyMqttTopic).(string)), "/")
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
		payload = _d[5]
		d = _d
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
		Type:       cmdMapForRx[msMsg.Command],
	}
	msgPL := msgml.NewData()
	msgPL.Value = payload

	messages = append(messages, msg)

	// update the payload details on return
	defer func() { msg.Payloads = []msgml.Data{msgPL} }()

	// verify node and sensor ids
	nID, err := strconv.ParseUint(msg.NodeID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Invalid node id: %s", msg.NodeID)
	}
	if nID > idBroadcastInt {
		return nil, fmt.Errorf("Invalid node id: %s", msg.NodeID)
	}
	sID, err := strconv.ParseUint(msg.SensorID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Invalid sensor id: %s", msg.SensorID)
	}
	if sID > idBroadcastInt {
		return nil, fmt.Errorf("Invalid sensor id: %s", msg.SensorID)
	}

	// Remove sensor id, if it is a internal message
	if msg.SensorID == idBroadcast {
		msg.SensorID = ""
	}

	// Remove node id, if it is a broadcast message
	if msg.NodeID == idBroadcast {
		msg.NodeID = ""
	}

	// set labels
	msgPL.Labels.Set(LabelType, msMsg.Type)
	if msg.NodeID != "" { // set node id if available
		msgPL.Labels.Set(LabelNodeID, msg.NodeID)
	}
	if msg.SensorID != "" { // set sensor id if available
		msgPL.Labels.Set(LabelSensorID, msg.SensorID)
	}

	// internal functions
	updateFieldData := func() error {
		_field, ok := setReqFieldMapForRx[msMsg.Type]
		if ok {
			msgPL.Labels.Set(LabelTypeString, _field)
		} else {
			_field = "V_CUSTOM"
			zap.L().Warn("This set, req not found. update this. Setting as V_CUSTOM", zap.Any("msMsg", msMsg))
		}

		// get type and unit
		if typeUnit, ok := metricTypeAndUnit[_field]; ok {
			msgPL.Name = _field
			msgPL.MetricType = typeUnit.Type
			msgPL.Unit = typeUnit.Unit
		} else {
			// not supported? should I have to return from here?
		}
		return nil
	}

	if msg.IsReceived && msg.IsAck { // process acknowledgement message
		// MySensors support ack message only for field/variable level, not for others
		if msMsg.SensorID != idBroadcast && (msMsg.Command == cmdSet || msMsg.Command == cmdRequest) {
			err := updateFieldData()
			if err != nil {
				return nil, err
			}
			return messages, nil
		} else if msMsg.NodeID != idBroadcast {
			if msMsg.Command == cmdInternal {
				switch msMsg.Type { // valid only for this list
				case typeInternalConfigResponse:
					msgPL.Name = "I_CONFIG"
				case typeInternalHeartBeatRequest:
					msgPL.Name = nml.ActionHeartbeatRequest
				case typeInternalIDResponse:
					msgPL.Name = "I_ID_REQUEST"
				case typeInternalPresentation:
					msgPL.Name = nml.ActionRefreshNodeInfo
				case typeInternalReboot:
					msgPL.Name = nml.ActionReboot
				case typeInternalTime:
					msgPL.Name = "I_TIME"
				default:
					// leave it, will fail at the end of root if
				}
				if msgPL.Name != "" {
					return messages, nil
				}
			} else if msMsg.Command == cmdStream {
				if _type, ok := streamTypeMapForRx[msMsg.Type]; ok {
					msg.Type = msgml.TypeAction
					msgPL.Name = _type
					return messages, nil
				}
			}
		}
		return messages, fmt.Errorf("For this message ack not implemented, rawMessage: %v", msMsg)
	}

	// entering into normal message processing
	switch {

	case msMsg.SensorID != idBroadcast: // perform sensor related stuff
		switch msMsg.Command {
		case cmdPresentation:
			if _type, ok := presentationTypeMapForRx[msMsg.Type]; ok {
				msgPL.Name = ml.FieldName
				msgPL.Labels.Set(LabelTypeString, _type)
			} else {
				// not supported? should I have to return from here?
			}

		case cmdSet, cmdRequest:
			err := updateFieldData()
			if err != nil {
				return nil, err
			}

		default:
			// not supported? should I have to return from here?
		}

	case msMsg.NodeID != idBroadcast: // perform node related stuff
		switch msMsg.Command {

		case cmdPresentation:
			msgPL.Labels.Set(ml.LabelNodeLibraryVersion, payload)
			if _type, ok := presentationTypeMapForRx[msMsg.Type]; ok {
				if _type == "S_ARDUINO_REPEATER_NODE" || _type == "S_ARDUINO_NODE" {
					// this is a node lib version data
					msgPL.Name = ml.FieldNone
					if _type == "S_ARDUINO_REPEATER_NODE" {
						msgPL.Labels.Set(LabelNodeType, "repeater")
					}
				} else {
					// return?
				}
			} else {
				// return?
			}
		case cmdInternal:
			if _type, ok := internalTypeMapForRx[msMsg.Type]; ok {
				if fieldName, ok := internalValidFields[_type]; ok {
					msgPL.Name = fieldName
					msg.Type = msgml.TypeSet

					if fieldName == ml.LabelNodeVersion {
						msgPL.Labels.Set(ml.LabelNodeVersion, payload)
					}

					if fieldName == ml.FieldLocked { // update locked reason
						msgPL.Others.Set(LabelLockedReason, payload, nil)
						msgPL.Value = "true"
					}
				} else {
					msg.Type = msgml.TypeAction
					msgPL.Name = _type

					// filter implemented requests
					_, found := util.FindItem(customValidActions, _type)
					if !found {
						return nil, fmt.Errorf("This internal message handling not implemented: %s", _type)
					}
				}
			} else {
				return nil, fmt.Errorf("Message internal type not found: %s", msMsg.Type)
			}

		case cmdStream:
			if _type, ok := streamTypeMapForRx[msMsg.Type]; ok {
				msg.Type = msgml.TypeAction
				msgPL.Name = _type

				// filter implemented requests
				_, found := util.FindItem(customValidActions, _type)
				if !found {
					return nil, fmt.Errorf("This stream message handling not implemented: %s", _type)
				}
			} else {
				return nil, fmt.Errorf("Message stream type not found: %s", msMsg.Type)
			}
		}

	case msMsg.NodeID == idBroadcast:
		if _type, ok := internalTypeMapForRx[msMsg.Type]; ok {
			msg.Type = msgml.TypeAction
			msgPL.Name = _type

			// filter implemented requests
			_, found := util.FindItem(customValidActions, _type)
			if !found {
				return nil, fmt.Errorf("This internal message handling not implemented: %s", _type)
			}
		} else {
			return nil, fmt.Errorf("Message internal type not found: %s", msMsg.Type)
		}

	default:
		// if none of the above
	}

	return messages, nil
}

// converts message to MySensors specific type
func (ms *message) toMySensorsRaw(isMQTT bool) string {
	// raw message format
	// node-id;child-sensor-id;command;ack;type;payload
	if ms.NodeID == "" {
		ms.NodeID = idBroadcast
	}
	if ms.SensorID == "" {
		ms.SensorID = idBroadcast
	}
	if isMQTT {
		return fmt.Sprintf("%s/%s/%s/%s/%s", ms.NodeID, ms.SensorID, ms.Command, ms.Ack, ms.Type)
	}
	return fmt.Sprintf("%s;%s;%s;%s;%s;%s\n", ms.NodeID, ms.SensorID, ms.Command, ms.Ack, ms.Type, ms.Payload)
}

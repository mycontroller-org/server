package mysensors

import (
	"errors"
	"fmt"
	"strings"

	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	"github.com/mycontroller-org/backend/v2/pkg/util"
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
		Payload:  msg.Payload,
	}

	// update broadcast id
	if msMsg.NodeID == "" {
		msMsg.NodeID = idBroadcast
	}
	if msMsg.SensorID == "" {
		msMsg.SensorID = idBroadcast
	}

	// init labels and others
	msg.Labels = msg.Labels.Init()
	msg.Others = msg.Others.Init()

	if msg.IsAckEnabled {
		msMsg.Ack = "1"
	}

	rawMsg := &msgml.RawMessage{Timestamp: msg.Timestamp}
	rawMsg.Others = rawMsg.Others.Init()

	// get command
	switch msg.Type {

	case msgml.TypeSet:
		msMsg.Command = cmdSet
		msMsg.Type = msg.Labels.Get(LabelType)

	case msgml.TypeRequest:
		msMsg.Command = cmdRequest
		msMsg.Type = msg.Labels.Get(LabelType)

	case msgml.TypeInternal, msgml.TypeStream:
		msMsg.Command = cmdInternal
		// call functions
		err := handleRequests(p.GWConfig, msg.FieldName, msg, &msMsg)
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
		rawMsg.Others.Set(gwpl.KeyTopic, []string{msMsg.toMySensorsRaw(true)}, nil)

	default:
		return nil, fmt.Errorf("This protocol not implemented: %s", p.GWConfig.Provider.ProtocolType)
	}
	return rawMsg, nil
}

// ToMessage converts to mc specific
func (p *Provider) ToMessage(rawMsg *msgml.RawMessage) (*msgml.Message, error) {
	// init others map
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
		Type:       cmdMapForRx[msMsg.Command],
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
	msg.Labels.Set(LabelType, msMsg.Type)
	if msg.NodeID != "" { // set node id if available
		msg.Labels.Set(LabelNodeID, msg.NodeID)
	}
	if msg.SensorID != "" { // set sensor id if available
		msg.Labels.Set(LabelSensorID, msg.SensorID)
	}

	// internal functions
	updateFieldData := func() error {
		_field, ok := setReqFieldMapForRx[msMsg.Type]
		if ok {
			msg.Labels.Set(LabelTypeString, _field)
		} else {
			_field = "V_CUSTOM"
			zap.L().Warn("This set, req not found. update this. Setting as V_CUSTOM", zap.Any("msMsg", msMsg))
		}

		// get type and unit
		if typeUnit, ok := metricTypeAndUnit[_field]; ok {
			msg.FieldName = _field
			msg.MetricType = typeUnit.Type
			msg.Unit = typeUnit.Unit
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
			return msg, nil
		} else if msMsg.NodeID != idBroadcast && msMsg.Command == cmdInternal {
			switch msMsg.Type { // valid only for this list
			case typeInternalConfigResponse:
				msg.FieldName = "I_CONFIG"
			case typeInternalHeartBeatRequest:
				msg.FieldName = nml.FuncHeartbeat
			case typeInternalIDResponse:
				msg.FieldName = "I_ID_REQUEST"
			case typeInternalPresentation:
				msg.FieldName = nml.FuncRefreshNodeInfo
			case typeInternalReboot:
				msg.FieldName = nml.FuncReboot
			case typeInternalTime:
				msg.FieldName = "I_TIME"
			default:
				// leave it, will fail at the end of root if
			}
			if msg.FieldName != "" {
				return msg, nil
			}
		} else if msMsg.Command == cmdStream {
			// TODO: for streaming
			return nil, errors.New("Streaming ack support not implemented")
		}

		return msg, fmt.Errorf("For this message ack not implemented, rawMessage: %v", msMsg)
	}

	// entering into normal message processing
	switch {

	case msMsg.SensorID != idBroadcast: // perform sensor related stuff
		switch msMsg.Command {
		case cmdPresentation:
			if _type, ok := presentationTypeMapForRx[msMsg.Type]; ok {
				msg.FieldName = fml.FieldName
				msg.Labels.Set(LabelTypeString, _type)
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
			msg.Others.Set(fml.FieldLibraryVersion, msg.Payload, nil) // set lib version
			if _type, ok := presentationTypeMapForRx[msMsg.Type]; ok {
				if _type == "S_ARDUINO_REPEATER_NODE" || _type == "S_ARDUINO_NODE" {
					// this is a node data
					msg.FieldName = fml.FieldName
					if _type == "S_ARDUINO_REPEATER_NODE" {
						msg.Labels.Set(LabelNodeType, "repeater")
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
					msg.FieldName = fieldName
					msg.Type = msgml.TypeSet
					if fieldName == fml.FieldLocked { // update locked reason
						msg.Others.Set(LabelLockedReason, msg.Payload, nil)
						msg.Payload = "true"
					}
				} else {
					msg.Type = msgml.TypeInternal
					msg.FieldName = _type

					// filter implemented requests
					_, found := util.FindItem(internalValidRequests, _type)
					if !found {
						return nil, fmt.Errorf("This internal message handling not implemented: %s", _type)
					}
				}
			} else {
				return nil, fmt.Errorf("Message internal type not found: %s", msMsg.Type)
			}

		case cmdStream:
			if _type, ok := streamTypeMapForRx[msMsg.Type]; ok {
				msg.Type = msgml.TypeStream
				msg.FieldName = _type

				// filter implemented requests
				_, found := util.FindItem(internalValidRequests, _type)
				if !found {
					return nil, fmt.Errorf("This stream message handling not implemented: %s", _type)
				}
			} else {
				return nil, fmt.Errorf("Message stream type not found: %s", msMsg.Type)
			}
		}
	default:
		// if none of the above
	}

	return msg, nil
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

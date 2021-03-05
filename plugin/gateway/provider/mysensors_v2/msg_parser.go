package mysensors

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gateway/protocol"
	mtsml "github.com/mycontroller-org/backend/v2/plugin/metrics"
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
			if mt.Type == mtsml.MetricTypeBinary {
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
		err := handleActions(p.GatewayConfig, payload.Name, msg, &msMsg)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("This command not implemented: %s", msg.Type)
	}

	if msMsg.Type == "" {
		return nil, errors.New("command type should not be empty")
	}

	// enable or disable acknowledgement
	msMsg.Ack = p.getAcknowledgementStatus(&msMsg)

	// create rawMessage
	switch p.ProtocolType {
	case gwpl.TypeSerial, gwpl.TypeEthernet:
		rawMsg.Data = []byte(msMsg.toMySensorsRaw(false))

	case gwpl.TypeMQTT:
		rawMsg.Data = []byte(msMsg.Payload)
		rawMsg.Others.Set(gwpl.KeyMqttTopic, []string{msMsg.toMySensorsRaw(true)}, nil)

	default:
		return nil, fmt.Errorf("protocol type not implemented: %s", p.ProtocolType)
	}

	// set acknowledgement request status
	if msMsg.Ack == "1" {
		rawMsg.AcknowledgeEnabled = true
	}

	// set id for raw message
	rawMsg.ID = generateMessageID(&msMsg)

	return rawMsg, nil
}

// ToMessage converts to mycontroller specific
func (p *Provider) ToMessage(rawMsg *msgml.RawMessage) ([]*msgml.Message, error) {
	messages := make([]*msgml.Message, 0)

	msMsg, err := p.decodeRawMessage(rawMsg)
	if err != nil {
		return nil, err
	}

	// if it is a acknowledgement message send it to acknowledgement topic and proceed further
	if msMsg.Ack == "1" && rawMsg.IsReceived {
		msgID := generateMessageID(msMsg)
		topicAck := mcbus.GetTopicPostRawMessageAcknowledgement(p.GatewayConfig.ID, msgID)
		err := mcbus.Publish(topicAck, "acknowledgement received.")
		if err != nil {
			zap.L().Error("failed post acknowledgement status", zap.String("gateway", p.GatewayConfig.ID), zap.String("topic", topicAck), zap.Error(err))
		}
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
	msgPL.Value = msMsg.Payload

	messages = append(messages, msg)

	// update the payload details on return
	includePayloads := func() { msg.Payloads = []msgml.Data{msgPL} }
	defer includePayloads()

	err = verifyAndUpdateNodeSensorIDs(msMsg, msg)
	if err != nil {
		return nil, err
	}

	// set labels
	msgPL.Labels.Set(LabelType, msMsg.Type)
	if msg.NodeID != "" { // set node id if available
		msgPL.Labels.Set(LabelNodeID, msg.NodeID)
	}
	if msg.SensorID != "" { // set sensor id if available
		msgPL.Labels.Set(LabelSensorID, msg.SensorID)
	}

	// entering into normal message processing
	switch {

	case msMsg.SensorID != "": // perform sensor related stuff
		switch msMsg.Command {
		case cmdPresentation:
			if _type, ok := presentationTypeMapForRx[msMsg.Type]; ok {
				msgPL.Name = ml.FieldName
				msgPL.Labels.Set(LabelTypeString, _type)
			} else {
				// not supported? should I have to return from here?
			}

		case cmdSet, cmdRequest:
			err := updateFieldAndUnit(msMsg, &msgPL)
			if err != nil {
				return nil, err
			}

		default:
			// not supported? should I have to return from here?
		}

	case msMsg.NodeID != "": // perform node related stuff
		switch msMsg.Command {

		case cmdPresentation:
			msgPL.Labels.Set(ml.LabelNodeLibraryVersion, msMsg.Payload)
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
			proceedFurther, err := updateNodeInternalMessages(msg, &msgPL, msMsg)
			if err != nil || !proceedFurther {
				return nil, err
			}

		case cmdStream:
			if typeName, ok := streamTypeMapForRx[msMsg.Type]; ok {
				// update the requested action
				_, isActionRequest := ut.FindItem(customValidActions, typeName)
				if !isActionRequest {
					// do not care about other types
					return nil, nil
				}
				msg.Type = msgml.TypeAction
				msgPL.Name = typeName
			} else {
				return nil, fmt.Errorf("Message stream type not found: %s", msMsg.Type)
			}
		}

	case msMsg.NodeID == "": // don't case about gateway broadcast message
		return nil, nil

	default:
		zap.L().Warn("This message not handled", zap.String("gateway", p.GatewayConfig.ID), zap.Any("rawMessage", rawMsg))
		return nil, nil
	}

	return messages, nil
}

// helper functions

// decodes raw message into message, which is local struct
func (p *Provider) decodeRawMessage(rawMsg *msgml.RawMessage) (*message, error) {
	d := make([]string, 0)
	payload := ""

	// decode message from gateway
	switch p.ProtocolType {
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
		return nil, fmt.Errorf("This type not implements. protocol type: %s", p.ProtocolType)
	}

	// MySensors message
	msMsg := &message{
		NodeID:   d[0],
		SensorID: d[1],
		Command:  d[2],
		Ack:      d[3],
		Type:     d[4],
		Payload:  payload,
	}
	return msMsg, nil
}

// verify node and sensor ids
func verifyAndUpdateNodeSensorIDs(msMsg *message, msg *msgml.Message) error {
	nID, err := strconv.ParseUint(msMsg.NodeID, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid node id: %s", msMsg.NodeID)
	}
	if nID > idBroadcastInt {
		return fmt.Errorf("Invalid node id: %s", msMsg.NodeID)
	}
	sID, err := strconv.ParseUint(msMsg.SensorID, 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid sensor id: %s", msMsg.SensorID)
	}
	if sID > idBroadcastInt {
		return fmt.Errorf("Invalid sensor id: %s", msMsg.SensorID)
	}

	// Remove sensor id, if it is a internal message
	if msMsg.SensorID == idBroadcast {
		msMsg.SensorID = ""
	}

	// Remove node id, if it is a broadcast message
	if msMsg.NodeID == idBroadcast {
		msMsg.NodeID = ""
	}

	// update node id and sensor id
	msg.SensorID = msMsg.SensorID
	msg.NodeID = msMsg.NodeID

	return nil
}

// update field name and unit
func updateFieldAndUnit(msMsg *message, msgPL *msgml.Data) error {
	fieldName, ok := setReqFieldMapForRx[msMsg.Type]
	if ok {
		msgPL.Labels.Set(LabelTypeString, fieldName)
	} else {
		fieldName = "V_CUSTOM"
		zap.L().Warn("This set, req not found. update this. Setting as V_CUSTOM", zap.Any("msMsg", msMsg))
	}

	// get type and unit
	if typeUnit, ok := metricTypeAndUnit[fieldName]; ok {
		msgPL.Name = fieldName
		msgPL.MetricType = typeUnit.Type
		msgPL.Unit = typeUnit.Unit
	} else {
		// not supported? should I have to return from here?
	}
	return nil
}

// updates node internal message data, return true, if further actions required
// returns false, if the message can be dropped
func updateNodeInternalMessages(msg *msgml.Message, msgPL *msgml.Data, msMsg *message) (bool, error) {
	if msMsg.Ack == "1" { // do not care about internal ack messages
		return false, nil
	}
	if typeName, ok := internalTypeMapForRx[msMsg.Type]; ok {
		// update the requested access
		_, isActionRequest := ut.FindItem(customValidActions, typeName)
		if isActionRequest {
			msg.Type = msgml.TypeAction
			msgPL.Name = typeName
			return true, nil
		}

		// verify it is valid field to update
		if fieldName, ok := internalValidFields[typeName]; ok {
			msgPL.Name = fieldName
			msg.Type = msgml.TypeSet

			if fieldName == ml.LabelNodeVersion {
				msgPL.Labels.Set(ml.LabelNodeVersion, msMsg.Payload)
			}

			if fieldName == ml.FieldLocked { // update locked reason
				msgPL.Others.Set(LabelLockedReason, msMsg.Payload, nil)
				msgPL.Value = "true"
			}
			return true, nil
		}

		// if non hits just return from here
		return false, nil
	}
	return false, fmt.Errorf("Message internal type not found: %s", msMsg.Type)
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

func (p *Provider) getAcknowledgementStatus(msMsg *message) string {
	if msMsg.NodeID == idBroadcast {
		return "0"
	}
	switch msMsg.Command {
	case cmdStream:
		if !p.Config.EnableStreamMessageAck {
			return "0"
		}
	case cmdInternal:
		if !p.Config.EnableInternalMessageAck {
			return "0"
		}
	default:
		return "1"
	}
	return "1"
}

func generateMessageID(msMsg *message) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s", msMsg.NodeID, msMsg.SensorID, msMsg.Command, msMsg.Ack, msMsg.Type)
}

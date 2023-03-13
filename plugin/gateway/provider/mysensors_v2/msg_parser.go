package mysensors

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	"go.uber.org/zap"
)

// toRawMessage converts to gateway specific
func (p *Provider) toRawMessage(msg *msgTY.Message) (*msgTY.RawMessage, error) {
	if len(msg.Payloads) == 0 {
		return nil, errors.New("there is no payload details on the message")
	}

	payload := msg.Payloads[0]

	msMsg := message{
		NodeID:   msg.NodeID,
		SensorID: msg.SourceID,
		Command:  "",
		Ack:      "0",
		Type:     "",
		Payload:  payload.Value.String(),
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

	rawMsg := &msgTY.RawMessage{Timestamp: msg.Timestamp}
	rawMsg.Others = rawMsg.Others.Init()

	// get command
	switch msg.Type {
	case msgTY.TypeSet:
		msMsg.Command = cmdSet
		msMsg.Type = payload.Labels.Get(LabelType)
		if msMsg.Type == "" {
			for k, v := range setReqFieldMapForRx {
				if v == payload.Key {
					msMsg.Type = k
					break
				}
			}
		}
		if mt, ok := metricTypeAndUnit[payload.Key]; ok {
			if mt.Type == metricTY.MetricTypeBinary {
				switch strings.ToLower(payload.Value.String()) {
				case "true", "on":
					msMsg.Payload = payloadON
				case "false", "off":
					msMsg.Payload = payloadOFF
				}
			}
		}

	case msgTY.TypeRequest:
		msMsg.Command = cmdRequest
		msMsg.Type = payload.Labels.Get(LabelType)

	case msgTY.TypeAction:
		msMsg.Command = cmdInternal
		// call functions
		err := p.handleActions(p.GatewayConfig, payload.Key, msg, &msMsg)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("this command not implemented: %s", msg.Type)
	}

	if msMsg.Type == "" {
		return nil, errors.New("command type should not be empty")
	}

	// enable or disable acknowledgement
	msMsg.Ack = p.getAcknowledgementStatus(&msMsg)

	// create rawMessage
	switch p.ProtocolType {
	case gwPtl.TypeSerial, gwPtl.TypeEthernet:
		rawMsg.Data = []byte(msMsg.toMySensorsRaw(false))

	case gwPtl.TypeMQTT:
		rawMsg.Data = []byte(msMsg.Payload)
		rawMsg.Others.Set(gwPtl.KeyMqttTopic, []string{msMsg.toMySensorsRaw(true)}, nil)

	default:
		return nil, fmt.Errorf("protocol type not implemented: %s", p.ProtocolType)
	}

	// set acknowledgement request status
	if msMsg.Ack == "1" {
		rawMsg.IsAckEnabled = true
	}

	// set id for raw message
	rawMsg.ID = generateMessageID(&msMsg)

	return rawMsg, nil
}

// ConvertToMessages converts to mycontroller specific format
func (p *Provider) ConvertToMessages(rawMsg *msgTY.RawMessage) ([]*msgTY.Message, error) {
	if rawMsg == nil {
		return nil, nil
	}
	messages := make([]*msgTY.Message, 0)

	msMsg, err := p.decodeRawMessage(rawMsg)
	if err != nil {
		return nil, err
	}

	if msMsg == nil {
		return nil, nil
	}

	// if it is a acknowledgement message send it to acknowledgement topic and proceed further
	if msMsg.Ack == "1" && rawMsg.IsReceived {
		msgID := generateMessageID(msMsg)
		topicAck := fmt.Sprintf("%s.%s", p.GatewayConfig.ID, msgID)
		err := p.bus.Publish(topicAck, "acknowledgement received")
		if err != nil {
			p.logger.Error("failed post acknowledgement status", zap.String("gateway", p.GatewayConfig.ID), zap.String("topic", topicAck), zap.Error(err))
		}
	}

	// Message
	msg := &msgTY.Message{
		NodeID:     msMsg.NodeID,
		SourceID:   msMsg.SensorID,
		IsAck:      msMsg.Ack == "1",
		IsReceived: true,
		Timestamp:  rawMsg.Timestamp,
		Type:       cmdMapForRx[msMsg.Command],
	}
	msgPL := msgTY.NewPayload()
	msgPL.SetValue(msMsg.Payload)

	messages = append(messages, msg)

	// update the payload details on return
	includePayloads := func() { msg.Payloads = []msgTY.Payload{msgPL} }
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
	if msg.SourceID != "" { // set source id if available
		msgPL.Labels.Set(LabelSensorID, msg.SourceID)
	}

	// entering into normal message processing
	switch {
	case msMsg.NodeID == "" && msMsg.Command == cmdInternal: // broadcast messages
		if typeName, ok := internalTypeMapForRx[msMsg.Type]; ok {
			// update the requested action
			_, isActionRequest := utils.FindItem(customValidActions, typeName)
			if !isActionRequest {
				// do not care about other types
				return nil, nil
			}
			msg.Type = msgTY.TypeAction
			msgPL.Key = typeName
		} else {
			return nil, fmt.Errorf("message internal(broadcast) type not found: %s", msMsg.Type)
		}

	case msMsg.SensorID != "": // perform sensor related stuff
		switch msMsg.Command {
		case cmdPresentation:
			if _type, ok := presentationTypeMapForRx[msMsg.Type]; ok {
				msgPL.Key = types.FieldName
				msgPL.Labels.Set(LabelTypeString, _type)
			}
			// else: not supported? should I have to return from here?

		case cmdSet, cmdRequest:
			err := p.updateFieldAndUnit(msMsg, &msgPL)
			if err != nil {
				return nil, err
			}

		default:
			// not supported? should I have to return from here?
		}

	case msMsg.NodeID != "": // perform node related stuff
		switch msMsg.Command {

		case cmdPresentation:
			msgPL.Labels.Set(types.LabelNodeLibraryVersion, msMsg.Payload)
			if _type, ok := presentationTypeMapForRx[msMsg.Type]; ok {
				if _type == "S_ARDUINO_REPEATER_NODE" || _type == "S_ARDUINO_NODE" {
					// this is a node lib version data
					msgPL.Key = types.FieldNone
					if _type == "S_ARDUINO_REPEATER_NODE" {
						msgPL.Labels.Set(LabelNodeType, "repeater")
					}
				}
				// else: return?
			}
			// else: return?

		case cmdInternal:
			proceedFurther, extraMessages, err := updateNodeInternalMessages(msg, &msgPL, msMsg)
			if err != nil || !proceedFurther {
				return nil, err
			}

			if len(extraMessages) > 0 {
				messages = append(messages, extraMessages...)
			}

		case cmdStream:
			if typeName, ok := streamTypeMapForRx[msMsg.Type]; ok {
				// update the requested action
				_, isActionRequest := utils.FindItem(customValidActions, typeName)
				if !isActionRequest {
					// do not care about other types
					return nil, nil
				}
				msg.Type = msgTY.TypeAction
				msgPL.Key = typeName
			} else {
				return nil, fmt.Errorf("message stream type not found: %s", msMsg.Type)
			}
		}

	case msMsg.NodeID == "": // don't case about gateway broadcast message
		return nil, nil

	default:
		p.logger.Warn("This message not handled", zap.String("gateway", p.GatewayConfig.ID), zap.Any("rawMessage", rawMsg))
		return nil, nil
	}

	return messages, nil
}

// helper functions

// decodes raw message into message, which is local struct
func (p *Provider) decodeRawMessage(rawMsg *msgTY.RawMessage) (*message, error) {
	var d []string
	payload := ""

	// decode message from gateway
	switch p.ProtocolType {
	case gwPtl.TypeMQTT:
		// topic/node-id/child-sensor-id/command/ack/type
		// out_rfm69/11/1/1/0/0
		rData := strings.Split(string(rawMsg.Others.Get(gwPtl.KeyMqttTopic).(string)), "/")
		if len(rData) < 5 {
			p.logger.Error("invalid message format", zap.Any("rawMessage", rawMsg))
			return nil, nil
		}
		d = rData[len(rData)-5:]
		payload = convertor.ToString(rawMsg.Data)
	case gwPtl.TypeSerial, gwPtl.TypeEthernet:
		// node-id;child-sensor-id;command;ack;type;payload
		_d := strings.Split(convertor.ToString(rawMsg.Data), ";")
		if len(_d) < 6 {
			p.logger.Error("invalid message format", zap.String("rawMessage", convertor.ToString(rawMsg.Data)))
			return nil, nil
		}
		payload = _d[5]
		d = _d
	// implement this one
	default:
		return nil, fmt.Errorf("this type not implements. protocol type: %s", p.ProtocolType)
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
func verifyAndUpdateNodeSensorIDs(msMsg *message, msg *msgTY.Message) error {
	nID, err := strconv.ParseUint(msMsg.NodeID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid node id: %s", msMsg.NodeID)
	}
	if nID > idBroadcastInt {
		return fmt.Errorf("invalid node id: %s", msMsg.NodeID)
	}
	sID, err := strconv.ParseUint(msMsg.SensorID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid sensor id: %s", msMsg.SensorID)
	}
	if sID > idBroadcastInt {
		return fmt.Errorf("invalid sensor id: %s", msMsg.SensorID)
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
	msg.SourceID = msMsg.SensorID
	msg.NodeID = msMsg.NodeID

	return nil
}

// update field name and unit
func (p *Provider) updateFieldAndUnit(msMsg *message, msgPL *msgTY.Payload) error {
	fieldName, ok := setReqFieldMapForRx[msMsg.Type]
	if ok {
		msgPL.Labels.Set(LabelTypeString, fieldName)
	} else {
		fieldName = "V_CUSTOM"
		p.logger.Warn("set or req not found. update this. Setting as V_CUSTOM", zap.Any("msMsg", msMsg))
	}

	// get type and unit
	if typeUnit, ok := metricTypeAndUnit[fieldName]; ok {
		msgPL.Key = fieldName
		msgPL.MetricType = typeUnit.Type
		msgPL.Unit = typeUnit.Unit
	}
	// else: not supported? should I have to return from here?

	return nil
}

// updates node internal message data, return true, if further actions required
// returns false, if the message can be dropped
func updateNodeInternalMessages(msg *msgTY.Message, msgPL *msgTY.Payload, msMsg *message) (bool, []*msgTY.Message, error) {
	if msMsg.Ack == "1" { // do not care about internal ack messages
		return false, nil, nil
	}
	if typeName, ok := internalTypeMapForRx[msMsg.Type]; ok {
		extraMessages := make([]*msgTY.Message, 0)
		// update the requested access
		_, isActionRequest := utils.FindItem(customValidActions, typeName)
		if isActionRequest {
			msg.Type = msgTY.TypeAction
			msgPL.Key = typeName
			return true, nil, nil
		}

		// verify it is valid field to update
		if fieldName, ok := internalValidFields[typeName]; ok {
			msgPL.Key = fieldName
			msg.Type = msgTY.TypeSet

			switch fieldName {
			case types.LabelNodeVersion:
				msgPL.Labels.Set(types.LabelNodeVersion, msMsg.Payload)

			case types.FieldLocked:
				msgPL.Others.Set(LabelLockedReason, msMsg.Payload, nil)
				msgPL.Value = "true"

			case LabelSmartSleepNode, types.FieldHeartbeat, types.FieldParentID:

				// mark it as sleeping node
				if fieldName == LabelSmartSleepNode {
					msgPL.Labels.Set(types.LabelNodeSleepNode, "true")
					msgPL.Labels.Set(LabelSmartSleepNode, "true")
				}

				// update for discover response
				if fieldName == types.FieldParentID {
					msgPL.Labels.Set(types.FieldParentID, msMsg.Payload)
					msgPL.Others.Set(types.FieldParentID, msMsg.Payload, nil)
				}

				switch typeName {
				case "I_PRE_SLEEP_NOTIFICATION", "I_HEARTBEAT_RESPONSE", "I_DISCOVER_RESPONSE":
					if typeName == "I_PRE_SLEEP_NOTIFICATION" {
						msgPL.Others.Set(FieldAwakeDuration, msMsg.Payload, nil)
					}
					// post an action to sen the messages in the queue
					awakeActionMsg := &msgTY.Message{
						GatewayID:  msg.GatewayID,
						NodeID:     msMsg.NodeID,
						IsAck:      false,
						IsReceived: true,
						Timestamp:  msg.Timestamp,
						Type:       msgTY.TypeAction,
						Payloads:   []msgTY.Payload{{Key: msgTY.TypeAction, Value: nodeTY.ActionAwake}},
					}
					extraMessages = append(extraMessages, awakeActionMsg)

				case "I_POST_SLEEP_NOTIFICATION":
					msgPL.Others.Set(FieldSleepDuration, msMsg.Payload, nil)

				}
			default:
				// noop
			}

			return true, extraMessages, nil
		}

		// if non hits just return from here
		return false, nil, nil
	}
	return false, nil, fmt.Errorf("message internal type not found: %s", msMsg.Type)
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

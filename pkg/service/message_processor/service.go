package messageprocessor

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	sourceAPI "github.com/mycontroller-org/backend/v2/pkg/api/source"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	fieldML "github.com/mycontroller-org/backend/v2/pkg/model/field"
	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	sourceML "github.com/mycontroller-org/backend/v2/pkg/model/source"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/pkg/utils/javascript"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	mtsML "github.com/mycontroller-org/backend/v2/plugin/metrics"
	"go.uber.org/zap"
)

var (
	msgQueue  *queueUtils.Queue
	queueSize = int(1000)
	workers   = 1
)

// Init message process engine
func Init() error {
	msgQueue = queueUtils.New("message_processor", queueSize, processMessage, workers)

	// on message receive add it in to our local queue
	_, err := mcbus.Subscribe(mcbus.GetTopicPostMessageToCore(), onMessageReceive)
	if err != nil {
		return err
	}

	return nil
}

func onMessageReceive(event *busML.BusData) {
	msg := &msgML.Message{}
	err := event.ToStruct(msg)
	if err != nil {
		zap.L().Warn("Failed to convet to target type", zap.Error(err))
		return
	}

	if msg == nil {
		zap.L().Warn("Received a nil message", zap.Any("event", event))
		return
	}
	zap.L().Debug("Message added into processing queue", zap.Any("message", msg))
	status := msgQueue.Produce(msg)
	if !status {
		zap.L().Warn("Failed to store the message into queue", zap.Any("message", msg))
	}
}

// Close message process engine
func Close() error {
	msgQueue.Close()
	return nil
}

// processMessage from the queue
func processMessage(item interface{}) {
	msg := item.(*msgML.Message)
	zap.L().Debug("Starting Message Processing", zap.Any("message", msg))

	switch {
	case msg.SourceID != "":
		switch msg.Type {
		case msgML.TypeSet: // set fields
			err := setFieldData(msg)
			if err != nil {
				zap.L().Error("error on field data set", zap.Error(err))
			}

		case msgML.TypeRequest: // request fields
			err := requestFieldData(msg)
			if err != nil {
				zap.L().Error("error on field data request", zap.Error(err))
			}

		case msgML.TypePresentation: // update source data, like name or other details
			err := updateSourceDetail(msg)
			if err != nil {
				zap.L().Error("error on source data update", zap.Error(err))
			}

		default:
			zap.L().Warn("message type not implemented for source", zap.String("type", msg.Type), zap.Any("message", msg))
		}

	case msg.NodeID != "":
		switch msg.Type {
		case msgML.TypeSet, msgML.TypePresentation: // set node specific data, like battery level, rssi, etc
			err := updateNodeData(msg)
			if err != nil {
				zap.L().Error("error on node data update", zap.Error(err))
			}

		case msgML.TypeRequest: // request node specific data

		case msgML.TypeAction: // providers will take care of action type messages
			clonedMsg := msg.Clone() // clone the message
			postMessage(clonedMsg)

		default:
			zap.L().Warn("message type not implemented for node", zap.String("type", msg.Type), zap.Any("message", msg))
		}

	case msg.NodeID == "" && msg.Type == msgML.TypeAction:
		clonedMsg := msg.Clone() // clone the message
		postMessage(clonedMsg)

	default:
		zap.L().Warn("this message not handled", zap.Any("message", msg))
	}

	zap.L().Debug("message processed", zap.String("timeTaken", time.Since(msg.Timestamp).String()), zap.Any("message", msg))
}

// update node detail
func updateNodeData(msg *msgML.Message) error {
	node, err := nodeAPI.GetByGatewayAndNodeID(msg.GatewayID, msg.NodeID)
	if err != nil { // TODO: check entry availability on error message
		node = &nodeML.Node{
			GatewayID: msg.GatewayID,
			NodeID:    msg.NodeID,
		}
	}

	// init labels and others
	node.Labels = node.Labels.Init()
	node.Others = node.Others.Init()

	// update last seen
	node.LastSeen = msg.Timestamp

	for _, d := range msg.Payloads {
		// update labels
		node.Labels.CopyFrom(d.Labels)

		switch d.Name { // set node name
		case model.FieldName:
			if !node.Labels.GetIgnoreBool(model.LabelName) {
				node.Name = d.Value
			}

		case model.FieldBatteryLevel: // set battery level
			// update battery level
			bl, err := strconv.ParseFloat(d.Value, 64)
			if err != nil {
				zap.L().Error("unable to parse batter level", zap.Error(err))
				return err
			}
			node.Others.Set(d.Name, bl, node.Labels)
			return writeNodeMetric(node, mtsML.MetricTypeGaugeFloat, model.FieldBatteryLevel, bl)

		default:
			if d.Name != model.FieldNone {
				node.Others.Set(d.Name, d.Value, node.Labels)
				// TODO: Do we need to report to metric strore?
			}

		}

		// update labels and Others
		node.Labels.CopyFrom(d.Labels)
		node.Others.CopyFrom(d.Others, node.Labels)
	}

	// save node data
	err = nodeAPI.Save(node)
	if err != nil {
		zap.L().Error("unable to update save the node data", zap.Error(err), zap.Any("node", node))
		return err
	}

	// post field data to event listeners
	postEvent(mcbus.TopicEventNode, node)

	return nil
}

func updateSourceDetail(msg *msgML.Message) error {
	source, err := sourceAPI.GetByIDs(msg.GatewayID, msg.NodeID, msg.SourceID)
	if err != nil { // TODO: check entry availability on error message
		source = &sourceML.Source{
			GatewayID: msg.GatewayID,
			NodeID:    msg.NodeID,
			SourceID:  msg.SourceID,
		}
	}

	// update last seen
	source.LastSeen = msg.Timestamp

	// init labels and others
	source.Labels = source.Labels.Init()
	source.Others = source.Others.Init()

	for _, payload := range msg.Payloads {
		switch payload.Name {
		case model.FieldName: // set name
			if !source.Labels.GetIgnoreBool(model.LabelName) {
				source.Name = payload.Value
			}

		default: // set other variables
			if payload.Name != model.FieldNone {
				source.Others.Set(payload.Name, payload.Value, source.Labels)
				// TODO: Do we need to report to metric strore?
			}
		}

		// update labels and Others
		source.Labels.CopyFrom(payload.Labels)
		source.Others.CopyFrom(payload.Others, source.Labels)
	}

	err = sourceAPI.Save(source)
	if err != nil {
		zap.L().Error("unable to update the source in to database", zap.Error(err), zap.Any("source", source))
		return err
	}
	// post field data to event listeners
	postEvent(mcbus.TopicEventSource, source)
	return nil
}

func setFieldData(msg *msgML.Message) error {
	for _, payload := range msg.Payloads {
		field, err := fieldAPI.GetByIDs(msg.GatewayID, msg.NodeID, msg.SourceID, payload.Name)
		if err != nil { // TODO: check entry availability on error message
			field = &fieldML.Field{
				GatewayID: msg.GatewayID,
				NodeID:    msg.NodeID,
				SourceID:  msg.SourceID,
				FieldID:   payload.Name,
			}
		}

		value := payload.Value

		// if custom payload formatter supplied
		if formatter := field.Formatter.OnReceive; msg.IsReceived && formatter != "" {
			startTime := time.Now()

			scriptInput := map[string]interface{}{
				"value":         payload.Value,
				"lastValue":     field.Current.Value,
				"previousValue": field.Previous.Value,
			}

			responseValue, err := javascript.Execute(formatter, scriptInput)
			if err != nil {
				zap.L().Error("error on executing script", zap.Error(err), zap.Any("inputValue", payload.Value), zap.String("gateway", field.GatewayID), zap.String("node", field.NodeID), zap.String("source", field.SourceID), zap.String("fieldID", field.FieldID), zap.String("script", formatter))
				return err
			}

			formattedValue := ""
			if responseValue == nil {
				zap.L().Error("returned nil value", zap.String("formatter", formatter))
				return errors.New("formatter returned nil value")
			}

			extraFields := map[string]interface{}{}
			if mapValue, ok := responseValue.(map[string]interface{}); ok {
				if _, found := mapValue["value"]; !found {
					zap.L().Error("value field not updated", zap.Any("received", mapValue), zap.String("formatter", formatter))
					return errors.New("formatter returned nil value")
				}
				for key, value := range mapValue {
					// if we see "value" key, update it on value
					// if we see others map update it on others map
					if key == model.KeyValue {
						formattedValue = utils.ToString(value)
					} else if key == model.KeyOthers {
						othersMap, ok := value.(map[string]interface{})
						if ok {
							for oKey, oValue := range othersMap {
								field.Others.Set(oKey, oValue, nil)
							}
						}
					} else {
						extraFields[key] = value
					}
				}
			} else {
				formattedValue = utils.ToString(responseValue)
			}

			// update the formatted value
			zap.L().Debug("formatting done", zap.Any("oldValue", payload.Value), zap.String("newValue", formattedValue), zap.String("timeTaken", time.Since(startTime).String()))

			// update formatted value into value
			value = formattedValue

			// update extra fields if any
			if len(extraFields) > 0 {
				labels := payload.Labels.Clone()
				err := updateExtraFieldsData(extraFields, msg, labels)
				if err != nil {
					zap.L().Error("error on updating extra fields", zap.Error(err))
				}
			}
		}
		err = updateFieldData(field, payload.Name, payload.Name, payload.MetricType, payload.Unit, payload.Labels, payload.Others, value, msg)
		if err != nil {
			zap.L().Error("error on updating field data", zap.Error(err), zap.String("gateway", msg.GatewayID), zap.String("node", msg.NodeID), zap.String("source", msg.SourceID), zap.String("field", payload.Name))
		}
	}
	return nil
}

func updateExtraFieldsData(extraFields map[string]interface{}, msg *msgML.Message, labels cmap.CustomStringMap) error {
	units := map[string]string{}
	metricTypes := map[string]string{}

	// update extraLabels
	if eLabels, ok := extraFields[model.KeyLabels]; ok {
		if extraLabels, ok := eLabels.(map[string]interface{}); ok {
			for key, val := range extraLabels {
				stringValue := utils.ToString(val)
				labels.Set(key, stringValue)
			}
		}
	}

	// update units
	if value, ok := extraFields[model.KeyUnits]; ok {
		if unitsRaw, ok := value.(map[string]interface{}); ok {
			for key, val := range unitsRaw {
				stringValue := utils.ToString(val)
				units[key] = stringValue
			}
		}
	}

	// update metricTypes
	if value, ok := extraFields[model.KeyMetricTypes]; ok {
		if mTypeRaw, ok := value.(map[string]interface{}); ok {
			for key, value := range mTypeRaw {
				stringValue := utils.ToString(value)
				metricTypes[key] = stringValue
			}
		}
	}

	// remove labels, units and metricTypes
	delete(extraFields, model.KeyLabels)
	delete(extraFields, model.KeyUnits)
	delete(extraFields, model.KeyMetricTypes)

	for id, value := range extraFields {
		fieldId := utils.ToString(id)
		metricType := mtsML.MetricTypeNone
		unit := ""
		// update metricType and unit
		if mType, ok := metricTypes[fieldId]; ok {
			metricType = mType
		}
		if unitString, ok := units[fieldId]; ok {
			unit = unitString
		}
		err := updateFieldData(nil, fieldId, fieldId, metricType, unit, labels, nil, value, msg)
		if err != nil {
			zap.L().Error("error on updating field data", zap.Error(err), zap.String("gateway", msg.GatewayID), zap.String("node", msg.NodeID), zap.String("source", msg.SourceID), zap.String("field", fieldId))
		}
	}

	return nil
}

func updateFieldData(
	field *fieldML.Field, fieldId, name, metricType, unit string, labels cmap.CustomStringMap,
	others cmap.CustomMap, value interface{}, msg *msgML.Message) error {

	if field == nil {
		updateField, err := fieldAPI.GetByIDs(msg.GatewayID, msg.NodeID, msg.SourceID, fieldId)
		if err != nil { // TODO: check entry availability on error message
			field = &fieldML.Field{
				GatewayID: msg.GatewayID,
				NodeID:    msg.NodeID,
				SourceID:  msg.SourceID,
				FieldID:   fieldId,
			}
			field.Labels = cmap.CustomStringMap{}
			field.Others = cmap.CustomMap{}
		} else {
			field = updateField
		}
	}

	// init labels and others
	labels = labels.Init()
	others = others.Init()

	// init field labels and others
	field.Labels = field.Labels.Init()
	field.Others = field.Others.Init()

	// update last seen
	field.LastSeen = msg.Timestamp

	// update name
	if !field.Labels.GetIgnoreBool(model.LabelName) {
		field.Name = name
	}

	// update type
	if !field.Labels.GetIgnoreBool(model.LabelMetricType) {
		field.MetricType = metricType
	}
	// update unit
	if !field.Labels.GetIgnoreBool(model.LabelUnit) {
		field.Unit = unit
	}

	zap.L().Debug("field", zap.Any("field", field))
	// update labels and others
	field.Labels.CopyFrom(labels)               // copy labels
	field.Others.CopyFrom(others, field.Labels) // copy other fields

	// convert value to specified metric type
	// convert payload to actual type
	var convertedValue interface{}
	switch field.MetricType {

	case mtsML.MetricTypeBinary:
		convertedValue = utils.ToBool(value)

	case mtsML.MetricTypeGaugeFloat:
		convertedValue = utils.ToFloat(value)

	case mtsML.MetricTypeGauge, mtsML.MetricTypeCounter:
		convertedValue = utils.ToInteger(value)

	case mtsML.MetricTypeNone:
		convertedValue = value

	case mtsML.MetricTypeString:
		convertedValue = utils.ToString(value)

	case mtsML.MetricTypeGEO: // Implement geo
		convertedValue = value

	default:
		zap.L().Error("unknown data type on a field", zap.Any("message", msg))
		return fmt.Errorf("unknown metricType: %s", field.MetricType)
	}

	// update shift old payload and update current payload
	field.Previous = field.Current
	field.Current = fieldML.Payload{Value: convertedValue, IsReceived: msg.IsReceived, Timestamp: msg.Timestamp}

	// update no change since
	oldValue := fmt.Sprintf("%v", field.Previous.Value)
	newValue := fmt.Sprintf("%v", field.Current.Value)
	if oldValue != newValue {
		field.NoChangeSince = msg.Timestamp
	}

	startTime := time.Now()
	err := fieldAPI.Save(field)
	if err != nil {
		zap.L().Error("failed to update field in to database", zap.Error(err), zap.Any("field", field))
	} else {
		zap.L().Debug("inserted in to storage db", zap.String("timeTaken", time.Since(startTime).String()))
	}

	// post field data to event listeners
	postEvent(mcbus.TopicEventFieldSet, field)

	startTime = time.Now()
	updateMetric := true
	if field.MetricType == mtsML.MetricTypeNone {
		updateMetric = false
	}
	// for binary do not update duplicate values
	if field.MetricType == mtsML.MetricTypeBinary {
		updateMetric = field.Current.Timestamp.Equal(field.NoChangeSince)
	}
	if updateMetric {
		err = writeFieldMetric(field)
		if err != nil {
			return err
		}
	} else {
		zap.L().Debug("skipped metric update", zap.Any("field", field))
	}
	return nil
}

func requestFieldData(msg *msgML.Message) error {
	payloads := make([]msgML.Data, 0)
	for _, payload := range msg.Payloads {
		field, err := fieldAPI.GetByIDs(msg.GatewayID, msg.NodeID, msg.SourceID, payload.Name)
		if err != nil {
			// TODO: check availability error message from storage
			continue
		}

		if field.Current.Value != nil {
			payload.Value = fmt.Sprintf("%v", field.Current.Value) // update payload
			if payload.Value != "" {                               // if the value is not empty update it
				payload.Labels = field.Labels.Clone()
				clonedData := payload.Clone() // clone the message
				payloads = append(payloads, clonedData)
			}
		}
		// post field data to event listeners
		// NOTE: if the entry not available in database request topic will not be sent
		postEvent(mcbus.TopicEventFieldRequest, field)
	}

	if len(payloads) > 0 {
		clonedMsg := msg.Clone()         // clone the message
		clonedMsg.Timestamp = time.Now() // set current timestamp
		clonedMsg.Payloads = payloads    // update payload
		clonedMsg.Type = msgML.TypeSet   // change type to set
		postMessage(clonedMsg)
	} else {
		zap.L().Debug("no data found for this request", zap.Any("message", msg))
	}
	return nil
}

// topic to send message to provider gateway
func postMessage(msg *msgML.Message) {
	if msg.IsAck {
		return // do not respond for ack message
	}
	topic := mcbus.GetTopicPostMessageToProvider(msg.GatewayID)
	msg.IsReceived = false
	err := mcbus.Publish(topic, msg)
	if err != nil {
		zap.L().Error("error on posting message", zap.String("topic", topic), zap.Any("message", msg), zap.Error(err))
	}
}

// sends updated resource as event.
func postEvent(eventTopic string, resource interface{}) {
	err := mcbus.Publish(mcbus.FormatTopic(eventTopic), resource)
	if err != nil {
		zap.L().Error("error on posting resource data", zap.String("topic", eventTopic), zap.Any("resource", resource), zap.Error(err))
	}
}

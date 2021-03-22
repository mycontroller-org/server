package messageprocessor

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	sensorAPI "github.com/mycontroller-org/backend/v2/pkg/api/sensor"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	busML "github.com/mycontroller-org/backend/v2/pkg/model/bus"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	sml "github.com/mycontroller-org/backend/v2/pkg/model/sensor"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	mts "github.com/mycontroller-org/backend/v2/pkg/service/metrics"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/pkg/utils/javascript"
	queueUtils "github.com/mycontroller-org/backend/v2/pkg/utils/queue"
	mtsml "github.com/mycontroller-org/backend/v2/plugin/metrics"
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
	msg := &msgml.Message{}
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
	msg := item.(*msgml.Message)
	zap.L().Debug("Starting Message Processing", zap.Any("message", msg))

	switch {
	case msg.SensorID != "":
		switch msg.Type {
		case msgml.TypeSet: // set fields
			setFieldData(msg)

		case msgml.TypeRequest: // request fields
			requestFieldData(msg)

		case msgml.TypePresentation: // update sensor data, like name or other details
			updateSensorDetail(msg)

		default:
			zap.L().Warn("Message type not implemented for sensor", zap.String("type", msg.Type), zap.Any("message", msg))
		}

	case msg.NodeID != "":
		switch msg.Type {
		case msgml.TypeSet, msgml.TypePresentation: // set node specific data, like battery level, rssi, etc
			updateNodeData(msg)

		case msgml.TypeRequest: // request node specific data

		case msgml.TypeAction: // providers will take care of action type messages
			clonedMsg := msg.Clone() // clone the message
			postMessage(clonedMsg)

		default:
			zap.L().Warn("Message type not implemented for node", zap.String("type", msg.Type), zap.Any("message", msg))
		}

	case msg.NodeID == "" && msg.Type == msgml.TypeAction:
		clonedMsg := msg.Clone() // clone the message
		postMessage(clonedMsg)

	default:
		zap.L().Warn("This message not handled", zap.Any("message", msg))
	}

	zap.L().Debug("Message processed", zap.String("timeTaken", time.Since(msg.Timestamp).String()), zap.Any("message", msg))
}

// update node detail
func updateNodeData(msg *msgml.Message) error {
	node, err := nodeAPI.GetByGatewayAndNodeID(msg.GatewayID, msg.NodeID)
	if err != nil { // TODO: check entry availability on error message
		node = &nml.Node{
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
				zap.L().Error("Unable to parse batter level", zap.Error(err))
				return err
			}
			node.Others.Set(d.Name, bl, node.Labels)
			// TODO: send it to metric store

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
		zap.L().Error("Unable to update save the node data", zap.Error(err), zap.Any("node", node))
		return err
	}

	// post field data to event listeners
	postEvent(mcbus.TopicEventNode, node)

	return nil
}

func updateSensorDetail(msg *msgml.Message) error {
	sensor, err := sensorAPI.GetByIDs(msg.GatewayID, msg.NodeID, msg.SensorID)
	if err != nil { // TODO: check entry availability on error message
		sensor = &sml.Sensor{
			GatewayID: msg.GatewayID,
			NodeID:    msg.NodeID,
			SensorID:  msg.SensorID,
		}
	}

	// update last seen
	sensor.LastSeen = msg.Timestamp

	// init labels and others
	sensor.Labels = sensor.Labels.Init()
	sensor.Others = sensor.Others.Init()

	for _, payload := range msg.Payloads {
		switch payload.Name {
		case model.FieldName: // set name
			if !sensor.Labels.GetIgnoreBool(model.LabelName) {
				sensor.Name = payload.Value
			}

		default: // set other variables
			if payload.Name != model.FieldNone {
				sensor.Others.Set(payload.Name, payload.Value, sensor.Labels)
				// TODO: Do we need to report to metric strore?
			}
		}

		// update labels and Others
		sensor.Labels.CopyFrom(payload.Labels)
		sensor.Others.CopyFrom(payload.Others, sensor.Labels)
	}

	err = sensorAPI.Save(sensor)
	if err != nil {
		zap.L().Error("Unable to update the sensor in to database", zap.Error(err), zap.Any("sensor", sensor))
		return err
	}
	// post field data to event listeners
	postEvent(mcbus.TopicEventSensor, sensor)
	return nil
}

func setFieldData(msg *msgml.Message) error {
	for _, payload := range msg.Payloads {
		field, err := fieldAPI.GetByIDs(msg.GatewayID, msg.NodeID, msg.SensorID, payload.Name)
		if err != nil { // TODO: check entry availability on error message
			field = &fml.Field{
				GatewayID: msg.GatewayID,
				NodeID:    msg.NodeID,
				SensorID:  msg.SensorID,
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
				zap.L().Error("error on executing script", zap.Error(err), zap.Any("inputValue", payload.Value), zap.String("gateway", field.GatewayID), zap.String("node", field.NodeID), zap.String("sensor", field.SensorID), zap.String("fieldID", field.FieldID), zap.String("script", formatter))
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
			zap.L().Debug("Formatting done", zap.Any("oldValue", payload.Value), zap.String("newValue", formattedValue), zap.String("timeTaken", time.Since(startTime).String()))

			// update formatted value into value
			value = formattedValue

			// update extra fields if any
			if len(extraFields) > 0 {
				labels := payload.Labels.Clone()
				updateExtraFieldsData(extraFields, msg, labels)
			}
		}
		err = updateFieldData(field, payload.Name, payload.Name, payload.MetricType, payload.Unit, payload.Labels, payload.Others, value, msg)
		if err != nil {
			zap.L().Error("error on updating field data", zap.Error(err), zap.String("gateway", msg.GatewayID), zap.String("node", msg.NodeID), zap.String("sensor", msg.SensorID), zap.String("field", payload.Name))
		}
	}
	return nil
}

func updateExtraFieldsData(extraFields map[string]interface{}, msg *msgml.Message, labels cmap.CustomStringMap) error {
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
		metricType := mtsml.MetricTypeNone
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
			zap.L().Error("error on updating field data", zap.Error(err), zap.String("gateway", msg.GatewayID), zap.String("node", msg.NodeID), zap.String("sensor", msg.SensorID), zap.String("field", fieldId))
		}
	}

	return nil
}

func updateFieldData(field *fml.Field, fieldId, name, metricType, unit string, labels cmap.CustomStringMap,
	others cmap.CustomMap, value interface{}, msg *msgml.Message) error {
	if field == nil {
		updateField, err := fieldAPI.GetByIDs(msg.GatewayID, msg.NodeID, msg.SensorID, fieldId)
		if err != nil { // TODO: check entry availability on error message
			field = &fml.Field{
				GatewayID: msg.GatewayID,
				NodeID:    msg.NodeID,
				SensorID:  msg.SensorID,
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

	// update labels and others
	field.Labels.CopyFrom(labels)               // copy labels
	field.Others.CopyFrom(others, field.Labels) // copy other fields

	// convert value to specified metric type
	// convert payload to actual type
	var convertedValue interface{}
	switch field.MetricType {

	case mtsml.MetricTypeBinary:
		convertedValue = utils.ToBool(value)

	case mtsml.MetricTypeGaugeFloat:
		convertedValue = utils.ToFloat(value)

	case mtsml.MetricTypeGauge, mtsml.MetricTypeCounter:
		convertedValue = utils.ToInteger(value)

	case mtsml.MetricTypeNone:
		convertedValue = value

	case mtsml.MetricTypeString:
		convertedValue = utils.ToString(value)

	case mtsml.MetricTypeGEO: // Implement geo
		convertedValue = value

	default:
		zap.L().Error("unknown data type on a field", zap.Any("message", msg))
		return fmt.Errorf("unknown metricType: %s", field.MetricType)
	}

	// update shift old payload and update current payload
	field.Previous = field.Current
	field.Current = fml.Payload{Value: convertedValue, IsReceived: msg.IsReceived, Timestamp: msg.Timestamp}

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
	postEvent(mcbus.TopicEventSensorFieldSet, field)

	startTime = time.Now()
	updateMetric := true
	if field.MetricType == mtsml.MetricTypeNone {
		updateMetric = false
	}
	// for binary do not update duplicate values
	if field.MetricType == mtsml.MetricTypeBinary {
		updateMetric = field.Current.Timestamp.Equal(field.NoChangeSince)
	}
	if updateMetric {
		err = mts.SVC.Write(field)
		if err != nil {
			zap.L().Error("failed to write into metrics database", zap.Error(err), zap.Any("field", field))
		} else {
			zap.L().Debug("inserted in to metric db", zap.String("timeTaken", time.Since(startTime).String()))
		}
	} else {
		zap.L().Debug("skipped metric update", zap.Any("field", field))
	}
	return nil
}

func requestFieldData(msg *msgml.Message) error {
	payloads := make([]msgml.Data, 0)
	for _, payload := range msg.Payloads {
		field, err := fieldAPI.GetByIDs(msg.GatewayID, msg.NodeID, msg.SensorID, payload.Name)
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
		postEvent(mcbus.TopicEventSensorFieldRequest, field)
	}

	if len(payloads) > 0 {
		clonedMsg := msg.Clone()         // clone the message
		clonedMsg.Timestamp = time.Now() // set current timestamp
		clonedMsg.Payloads = payloads    // update payload
		clonedMsg.Type = msgml.TypeSet   // change type to set
		postMessage(clonedMsg)
	} else {
		zap.L().Debug("no data found for this request", zap.Any("message", msg))
	}
	return nil
}

// topic to send message to provider gateway
func postMessage(msg *msgml.Message) {
	if msg.IsAck {
		return // do not respond for ack message
	}
	topic := mcbus.GetTopicPostMessageToProvider(msg.GatewayID)
	msg.IsReceived = false
	err := mcbus.Publish(topic, msg)
	if err != nil {
		zap.L().Error("Error on posting message", zap.String("topic", topic), zap.Any("message", msg), zap.Error(err))
	}
}

// sends updated resource as event.
func postEvent(eventTopic string, resource interface{}) {
	err := mcbus.Publish(mcbus.FormatTopic(eventTopic), resource)
	if err != nil {
		zap.L().Error("Error on posting resource data", zap.String("topic", eventTopic), zap.Any("resource", resource), zap.Error(err))
	}
}

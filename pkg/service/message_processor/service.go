package messageprocessor

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	sensorAPI "github.com/mycontroller-org/backend/v2/pkg/api/sensor"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	"github.com/mycontroller-org/backend/v2/pkg/model/event"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	sml "github.com/mycontroller-org/backend/v2/pkg/model/sensor"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	mts "github.com/mycontroller-org/backend/v2/pkg/service/metrics"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	mtsml "github.com/mycontroller-org/backend/v2/plugin/metrics"
	"github.com/robertkrimen/otto"
	"go.uber.org/zap"
)

var (
	msgQueue  *q.BoundedQueue
	queueSize = int(1000)
)

// Init message process engine
func Init() error {
	msgQueue = utils.GetQueue("message_processor", queueSize)

	// on message receive add it in to our local queue
	_, err := mcbus.Subscribe(mcbus.GetTopicPostMessageToCore(), onMessageReceive)
	if err != nil {
		return err
	}

	msgQueue.StartConsumers(1, processMessage)
	return nil
}

func onMessageReceive(event *event.Event) {
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
	msgQueue.Stop()
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
		case ml.FieldName:
			if !node.Labels.GetIgnoreBool(ml.LabelName) {
				node.Name = d.Value
			}

		case ml.FieldBatteryLevel: // set battery level
			// update battery level
			bl, err := strconv.ParseFloat(d.Value, 64)
			if err != nil {
				zap.L().Error("Unable to parse batter level", zap.Error(err))
				return err
			}
			node.Others.Set(d.Name, bl, node.Labels)
			// TODO: send it to metric store

		default:
			if d.Name != ml.FieldNone {
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
		case ml.FieldName: // set name
			if !sensor.Labels.GetIgnoreBool(ml.LabelName) {
				sensor.Name = payload.Value
			}

		default: // set other variables
			if payload.Name != ml.FieldNone {
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
		// current payload
		currentPayload := fml.Payload{}
		var err error
		var pl interface{}

		// convert payload to actual type
		switch payload.MetricType {

		case mtsml.MetricTypeBinary:
			pl, err = strconv.ParseBool(payload.Value)

		case mtsml.MetricTypeGaugeFloat:
			pl, err = strconv.ParseFloat(payload.Value, 64)

		case mtsml.MetricTypeGauge, mtsml.MetricTypeCounter:
			pl, err = strconv.ParseInt(payload.Value, 10, 64)

		case mtsml.MetricTypeNone:
			pl = payload.Value

		case mtsml.MetricTypeGEO: // Implement geo
			pl = payload.Value

		default:
			zap.L().Error("Unknown data type on a field", zap.Any("message", msg))
			return errors.New("Unknown data type on a field")
		}

		if err != nil {
			zap.L().Error("Unable to convert the payload to actual type", zap.Error(err), zap.Any("message", msg))
			return err
		}

		field, err := fieldAPI.GetByIDs(msg.GatewayID, msg.NodeID, msg.SensorID, payload.Name)
		if err != nil { // TODO: check entry availability on error message
			field = &fml.Field{
				GatewayID: msg.GatewayID,
				NodeID:    msg.NodeID,
				SensorID:  msg.SensorID,
				FieldID:   payload.Name,
			}
		}

		// update payload
		currentPayload = fml.Payload{Value: pl, IsReceived: msg.IsReceived, Timestamp: msg.Timestamp}

		// if custom payload formatter supplied
		if formatter := field.PayloadFormatter.OnReceive; msg.IsReceived && formatter != "" {
			startTime := time.Now()
			ottoVM := otto.New()
			ottoVM.Set("value", currentPayload.Value)

			ottoValue, err := ottoVM.Run(formatter)
			if err != nil {
				zap.L().Error("Failure on payload formatter", zap.String("formatter", formatter), zap.Error(err))
				return err
			}
			value, err := ottoValue.ToString()
			if err != nil {
				zap.L().Error("Failed to get value", zap.String("formatter", formatter), zap.Error(err))
				return err
			}
			// update the formatted value
			currentPayload.Value = value
			zap.L().Debug("Formatting done", zap.Any("oldValue", currentPayload.Value), zap.String("newValue", value), zap.String("timeTaken", time.Since(startTime).String()))
		}

		// update last seen
		field.LastSeen = msg.Timestamp

		// init labels and others
		field.Labels = field.Labels.Init()
		field.Others = field.Others.Init()

		// update name
		if !field.Labels.GetIgnoreBool(ml.LabelName) {
			field.Name = payload.Name
		}

		// update type
		if !field.Labels.GetIgnoreBool(ml.LabelMetricType) {
			field.MetricType = payload.MetricType
		}
		// update unit
		if !field.Labels.GetIgnoreBool(ml.LabelUnit) {
			field.Unit = payload.Unit
		}

		// update labels and others
		field.Labels.CopyFrom(payload.Labels)               // copy labels
		field.Others.CopyFrom(payload.Others, field.Labels) // copy other fields

		// update no change since
		oldValue := fmt.Sprintf("%v", field.Payload.Value)
		newValue := fmt.Sprintf("%v", currentPayload.Value)
		if oldValue != newValue {
			field.NoChangeSince = currentPayload.Timestamp
		}

		// update shift old payload and update current payload
		field.PreviousPayload = field.Payload
		field.Payload = currentPayload

		startTime := time.Now()
		err = fieldAPI.Save(field)
		if err != nil {
			zap.L().Error("Failed to update field in to database", zap.Error(err), zap.Any("field", field))
		} else {
			zap.L().Debug("Inserted in to storage db", zap.String("timeTaken", time.Since(startTime).String()))
		}

		// post field data to event listeners
		postEvent(mcbus.TopicEventSensorFieldSet, field)

		startTime = time.Now()
		updateMetric := true
		// for binary do not update duplicate values
		if field.MetricType == mtsml.MetricTypeBinary {
			updateMetric = field.Payload.Timestamp.Equal(field.NoChangeSince)
		}
		if updateMetric {
			err = mts.SVC.Write(field)
			if err != nil {
				zap.L().Error("Failed to write into metrics database", zap.Error(err), zap.Any("field", field))
			} else {
				zap.L().Debug("Inserted in to metric db", zap.String("timeTaken", time.Since(startTime).String()))
			}
		} else {
			zap.L().Debug("Skipped metric update", zap.Any("field", field))
		}

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

		if field.Payload.Value != nil {
			payload.Value = fmt.Sprintf("%v", field.Payload.Value) // update payload
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

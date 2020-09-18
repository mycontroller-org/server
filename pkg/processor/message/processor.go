package messageprocessor

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/mustafaturan/bus"
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	sensorAPI "github.com/mycontroller-org/backend/v2/pkg/api/sensor"
	"github.com/mycontroller-org/backend/v2/pkg/mcbus"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	mtrml "github.com/mycontroller-org/backend/v2/pkg/model/metric"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	sml "github.com/mycontroller-org/backend/v2/pkg/model/sensor"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	"go.uber.org/zap"
)

var (
	mq   *q.BoundedQueue
	size = int(1000)
)

// Init message process engine
func Init() error {
	mq = q.NewBoundedQueue(size, func(item interface{}) {
		zap.L().Error("Dropping an item, queue full", zap.Int("size", size), zap.Any("item", item))
	})

	onMessageReceive := func(e *bus.Event) {
		msg := e.Data.(*msgml.Message)
		if msg == nil {
			zap.L().Warn("Received a nil message", zap.Any("event", e))
			return
		}
		zap.L().Debug("Message added into processing queue", zap.Any("message", msg))
		status := mq.Produce(msg)
		if !status {
			zap.L().Warn("Failed to store the message into queue", zap.Any("message", msg))
		}
	}

	// on message receive add it in to our local queue
	mcbus.Subscribe(mcbus.TopicMsgFromGW, &bus.Handler{
		Matcher: mcbus.TopicMsgFromGW,
		Handle:  onMessageReceive,
	})

	mq.StartConsumers(1, processMessage)
	return nil
}

// Close message process engine
func Close() error {
	mq.Stop()
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
	node, err := nodeAPI.GetByIDs(msg.GatewayID, msg.NodeID)
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
	nodeAPI.Save(node)
	if err != nil {
		zap.L().Error("Unable to update save the node data", zap.Error(err), zap.Any("node", node))
		return err
	}

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

	for _, d := range msg.Payloads {
		switch d.Name {
		case ml.FieldName: // set name
			if !sensor.Labels.GetIgnoreBool(ml.LabelName) {
				sensor.Name = d.Value
			}

		default: // set other variables
			if d.Name != ml.FieldNone {
				sensor.Others.Set(d.Name, d.Value, sensor.Labels)
				// TODO: Do we need to report to metric strore?
			}
		}

		// update labels and Others
		sensor.Labels.CopyFrom(d.Labels)
		sensor.Others.CopyFrom(d.Others, sensor.Labels)
	}

	err = sensorAPI.Save(sensor)
	if err != nil {
		zap.L().Error("Unable to update the sensor in to database", zap.Error(err), zap.Any("sensor", sensor))
		return err
	}
	return nil
}

func setFieldData(msg *msgml.Message) error {
	for _, d := range msg.Payloads {
		// current payload
		cPL := fml.Payload{}
		var err error
		var pl interface{}

		// convert payload to actual type
		switch d.MetricType {

		case mtrml.MetricTypeBinary:
			pl, err = strconv.ParseBool(d.Value)

		case mtrml.MetricTypeGaugeFloat:
			pl, err = strconv.ParseFloat(d.Value, 64)

		case mtrml.MetricTypeGauge, mtrml.MetricTypeCounter:
			pl, err = strconv.ParseInt(d.Value, 10, 64)

		case mtrml.MetricTypeNone:
			pl = d.Value

		case mtrml.MetricTypeGEO: // Implement geo
			pl = d.Value

		default:
			zap.L().Error("Unknown data type on a field", zap.Any("message", msg))
			return errors.New("Unknown data type on a field")
		}

		if err != nil {
			zap.L().Error("Unable to convert the payload to actual type", zap.Error(err), zap.Any("message", msg))
			return err
		}

		// update payload
		cPL = fml.Payload{Value: pl, IsReceived: msg.IsReceived, Timestamp: msg.Timestamp}

		field, err := fieldAPI.GetByIDs(msg.GatewayID, msg.NodeID, msg.SensorID, d.Name)
		if err != nil { // TODO: check entry availability on error message
			field = &fml.Field{
				GatewayID: msg.GatewayID,
				NodeID:    msg.NodeID,
				SensorID:  msg.SensorID,
				FieldID:   d.Name,
			}
		}

		// update last seen
		field.LastSeen = msg.Timestamp

		// init labels and others
		field.Labels = field.Labels.Init()
		field.Others = field.Others.Init()

		// update name
		if !field.Labels.GetIgnoreBool(ml.LabelName) {
			field.Name = d.Name
		}

		// update type
		if !field.Labels.GetIgnoreBool(ml.LabelMetricType) {
			field.MetricType = d.MetricType
		}
		// update unit
		if !field.Labels.GetIgnoreBool(ml.LabelUnit) {
			field.Unit = d.Unit
		}

		// update labels and others
		field.Labels.CopyFrom(d.Labels)               // copy labels
		field.Others.CopyFrom(d.Others, field.Labels) // copy other fields

		// update shift old payload and update current payload
		field.PreviousPayload = field.Payload
		field.Payload = cPL

		start := time.Now()
		err = fieldAPI.Save(field)
		if err != nil {
			zap.L().Error("Failed to update field in to database", zap.Error(err), zap.Any("field", field))
		} else {
			zap.L().Debug("Inserted in to storage db", zap.String("timeTaken", time.Since(start).String()))
		}

		start = time.Now()
		err = svc.MTS.Write(field)
		if err != nil {
			zap.L().Error("Failed to write into metrics database", zap.Error(err), zap.Any("field", field))
		} else {
			zap.L().Debug("Inserted in to metric db", zap.String("timeTaken", time.Since(start).String()))
		}
	}
	return nil
}

func requestFieldData(msg *msgml.Message) error {
	payloads := make([]msgml.Data, 0)
	for _, d := range msg.Payloads {
		field, err := fieldAPI.GetByIDs(msg.GatewayID, msg.NodeID, msg.SensorID, d.Name)
		if err != nil { // TODO: check entry availability on error message
			continue
		}

		if field.Payload.Value != nil {
			clonedData := d.Clone()                          // clone the message
			d.Value = fmt.Sprintf("%v", field.Payload.Value) // update payload
			d.Labels = field.Labels.Clone()
			payloads = append(payloads, clonedData)
		}
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

func postMessage(msg *msgml.Message) {
	// topic to send message to gateway
	topic := fmt.Sprintf("%s_%s", mcbus.TopicMsg2GW, msg.GatewayID)
	msg.IsReceived = false
	mcbus.Publish(topic, msg)
}

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
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
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
	mcbus.Subscribe(mcbus.TopicMsgFromGW, &bus.Handler{
		Matcher: mcbus.TopicMsgFromGW,
		Handle:  onMessageReceive,
	})
	return nil
}

// Close message process engine
func Close() error {
	mq.Stop()
	return nil
}

// onMessageReceive from gateways
func onMessageReceive(e *bus.Event) {
	msg := e.Data.(*msgml.Message)
	if msg == nil {
		zap.L().Warn("Received a nil message", zap.Any("event", e))
		return
	}
	zap.L().Debug("message received", zap.Any("message", msg))

	switch {
	case msg.SensorID != "":
		switch msg.Type {
		case msgml.TypeSet, msgml.TypeRequest: // set a field
			isRequest := msg.Type == msgml.TypeRequest
			setReqFieldData(msg, isRequest)

		case msgml.TypePresentation: // update sensor data
			// update sensor name or properties
			updateSensorDetail(msg)

		default:
			zap.L().Warn("Message type not implemented for sensor", zap.String("type", msg.Type), zap.Any("message", msg))
		}

	case msg.NodeID != "":
		switch msg.Type {
		case msgml.TypeSet, msgml.TypePresentation: // set node specific data, like battery level, rssi, etc
			updateNodeData(msg)

		case msgml.TypeRequest: // request node specific data

		case msgml.TypeInternal, msgml.TypeStream: // providers will take care of internal and stream type messages
			clonedMsg := msg.Clone() // clone the message
			postMessage(clonedMsg)

		default:
			zap.L().Warn("Message type not implemented for node", zap.String("type", msg.Type), zap.Any("message", msg))
		}

	case msg.NodeID == "" && msg.Type == msgml.TypeInternal:
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

	switch msg.FieldName { // set node name
	case fml.FieldName:
		if !node.Labels.GetBool(ml.LabelIgnoreName) {
			node.Name = msg.Payload
		}

	case fml.FieldBatteryLevel: // set battery level
		// update battery level
		bl, err := strconv.ParseFloat(msg.Payload, 64)
		if err != nil {
			zap.L().Error("Unable to parse batter level", zap.Error(err))
			return err
		}
		node.Others.Set(msg.FieldName, bl, node.Labels)
		// send it to metric store

	default:
		node.Others.Set(msg.FieldName, msg.Payload, node.Labels)
	}

	// update labels and Others
	node.Labels.CopyFrom(msg.Labels)
	node.Others.CopyFrom(msg.Others, node.Labels)

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

	switch msg.FieldName { // set name
	case fml.FieldName:
		if !sensor.Labels.GetBool(ml.LabelIgnoreName) {
			sensor.Name = msg.Payload
		}

	default: // set other variables
		sensor.Others.Set(msg.FieldName, msg.Payload, sensor.Labels)
	}

	// update labels and Others
	sensor.Labels.CopyFrom(msg.Labels)
	sensor.Others.CopyFrom(msg.Others, sensor.Labels)

	err = sensorAPI.Save(sensor)
	if err != nil {
		zap.L().Error("Unable to update the sensor in to database", zap.Error(err), zap.Any("sensor", sensor))
		return err
	}
	return nil
}

func setReqFieldData(msg *msgml.Message, isRequest bool) error {
	// current payload
	cPL := fml.Payload{}

	if !isRequest { // parse payload, only for set type
		var err error
		var pl interface{}
		// convert payload to actual type
		switch msg.MetricType {

		case fml.MetricTypeBinary:
			pl, err = strconv.ParseBool(msg.Payload)

		case fml.MetricTypeGaugeFloat:
			pl, err = strconv.ParseFloat(msg.Payload, 64)

		case fml.MetricTypeGauge, fml.MetricTypeCounter:
			pl, err = strconv.ParseInt(msg.Payload, 10, 64)

		case fml.MetricTypeNone:
			pl = msg.Payload

		case fml.MetricTypeGEO: // Implement geo
			pl = msg.Payload

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
	}

	field, err := fieldAPI.GetByIDs(msg.GatewayID, msg.NodeID, msg.SensorID, msg.FieldName)
	if err != nil { // TODO: check entry availability on error message
		field = &fml.Field{
			GatewayID: msg.GatewayID,
			NodeID:    msg.NodeID,
			SensorID:  msg.SensorID,
			FieldID:   msg.FieldName,
		}
	}

	// update last seen
	field.LastSeen = msg.Timestamp

	// init labels and others
	field.Labels = field.Labels.Init()
	field.Others = field.Others.Init()

	// update name
	if !field.Labels.GetBool(cmap.GetIgnoreKey(fml.FieldName)) {
		field.Name = msg.FieldName
	}

	// update type
	if !field.Labels.GetBool(cmap.GetIgnoreKey(fml.FieldType)) {
		field.MetricType = msg.MetricType
	}
	// update unit
	if !field.Labels.GetBool(cmap.GetIgnoreKey(fml.FieldUnit)) {
		field.Unit = msg.Unit
	}

	// TODO: update labels and others
	field.Labels.CopyFrom(msg.Labels)
	field.Others.CopyFrom(msg.Others, field.Labels)

	if isRequest { // execute for request message
		if field.Payload.Value != nil {
			clonedMsg := msg.Clone()                                   // clone the message
			clonedMsg.Timestamp = time.Now()                           // set current timestamp
			clonedMsg.Payload = fmt.Sprintf("%v", field.Payload.Value) // update payload
			clonedMsg.Type = msgml.TypeSet                             // change type to set
			postMessage(clonedMsg)
		}
	} else {
		// update shift old payload and update current payload
		field.PreviousPayload = field.Payload
		field.Payload = cPL
	}

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
	return nil
}

func postMessage(msg *msgml.Message) {
	// topic to send message to gateway
	topic := fmt.Sprintf("%s_%s", mcbus.TopicMsg2GW, msg.GatewayID)
	msg.IsReceived = false
	mcbus.Publish(topic, msg)
}

package messageprocessor

import (
	"fmt"
	"strconv"
	"time"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/mustafaturan/bus"
	"github.com/mycontroller-org/backend/pkg/mcbus"
	ml "github.com/mycontroller-org/backend/pkg/model"
	msg "github.com/mycontroller-org/backend/pkg/model/message"
	nml "github.com/mycontroller-org/backend/pkg/model/node"
	sml "github.com/mycontroller-org/backend/pkg/model/sensor"
	svc "github.com/mycontroller-org/backend/pkg/service"
	"github.com/mycontroller-org/backend/pkg/util"
	ut "github.com/mycontroller-org/backend/pkg/util"
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
	m := e.Data.(*msg.Message)
	if m == nil {
		zap.L().Warn("Received a nil message", zap.Any("event", e))
		return
	}

	updateNodeLastSeen := true

	// update field value or sensor data
	if m.SensorID != "" {
		switch m.Command {
		case msg.CommandSet:
			updateSensorFieldData(m)
		case msg.CommandRequest:
			// requestSensorFieldData(s, m)
		case msg.CommandPresentation:
			// update sensor name or properties
			updateSensorData(m)
		default:
			zap.L().Warn("Found not implemented command for sensor", zap.Any("message", m))
		}
	}
	// update node data
	if m.NodeID != "" {
		switch m.Command {
		case msg.CommandSet:
		case msg.CommandRequest:
		case msg.CommandPresentation:
			updateNodeData(m)
		case msg.CommandInternal:
		case msg.CommandStream:
		default:
		}
		updateNodeLastSeen = false
	}
	// update node last seen
	if updateNodeLastSeen {

	}
	// update gateway last message

	zap.L().Debug("Message process completed", zap.String("timeTaken", time.Since(m.Timestamp).String()), zap.Any("message", m))
}

func updateNodeData(m *msg.Message) error {
	n := &nml.Node{
		ID:        fmt.Sprintf("%s_%s", m.GatewayID, m.NodeID),
		GatewayID: m.GatewayID,
		ShortID:   m.NodeID,
		LastSeen:  m.Timestamp,
	}
	f := []ml.Filter{
		{Key: "id", Operator: "eq", Value: n.ID},
	}
	node := &nml.Node{}
	err := svc.STG.FindOne(ml.EntityNode, f, node)
	if err != nil {
		node = n
		err = svc.STG.Insert(ml.EntityNode, node)
		if err != nil {
			zap.L().Error("Unable to insert the node in to database", zap.Error(err), zap.Any("node", node))
			return err
		}
	}
	if node.Others == nil {
		node.Others = map[string]interface{}{}
	}
	switch m.SubCommand {
	case msg.SubCmdName:
		if util.GetMapValue(node.Config, ml.CFGUpdateName, true).(bool) {
			node.Name = m.Payload
		}
	case msg.SubCmdBatteryLevel:
		// update battery level
		bl, err := strconv.ParseFloat(m.Payload, 64)
		if err != nil {
			zap.L().Error("Unable to parse batter level", zap.Error(err))
		}
		node.Others[m.Field] = bl
		// send it to metric store
	case msg.SubCmdDiscover, msg.SubCmdHeartbeat, msg.SubCmdPing:
		// TODO:
	default:
		node.Others[m.Field] = m.Payload
	}
	if len(m.Others) > 0 {
		ut.JoinMap(node.Others, m.Others)
	}
	err = svc.STG.Upsert(ml.EntityNode, f, node)
	if err != nil {
		zap.L().Error("Unable to update the node in to database", zap.Error(err), zap.Any("node", node))
	}
	return nil
}

func updateSensorData(m *msg.Message) error {
	s := &sml.Sensor{
		ID:        fmt.Sprintf("%s_%s_%s", m.GatewayID, m.NodeID, m.SensorID),
		GatewayID: m.GatewayID,
		NodeID:    m.NodeID,
		ShortID:   m.SensorID,
		LastSeen:  m.Timestamp,
	}
	f := []ml.Filter{
		{Key: "id", Operator: "eq", Value: s.ID},
	}
	sensor := &sml.Sensor{}
	err := svc.STG.FindOne(ml.EntitySensor, f, sensor)
	if err != nil {
		sensor = s
		err = svc.STG.Insert(ml.EntitySensor, sensor)
		if err != nil {
			zap.L().Error("Unable to insert the sensor in to database", zap.Error(err), zap.Any("sensor", sensor))
			return err
		}
	}
	if sensor.Others == nil {
		sensor.Others = map[string]interface{}{}
	}

	// update sensor name
	if m.SubCommand == msg.SubCmdName {
		if util.GetMapValue(sensor.Config, ml.CFGUpdateName, true).(bool) {
			sensor.Name = m.Payload
		}
	}
	if len(m.Others) > 0 {
		ut.JoinMap(sensor.Others, m.Others)
	}
	err = svc.STG.Upsert(ml.EntitySensor, f, sensor)
	if err != nil {
		zap.L().Error("Unable to update the sensor in to database", zap.Error(err), zap.Any("sensor", sensor))
		return err
	}
	return nil
}

func updateSensorFieldData(m *msg.Message) error {
	var err error
	var pl interface{}
	// convert payload to actual type
	switch m.PayloadType {
	case msg.PayloadTypeBoolean:
		pl, err = strconv.ParseBool(m.Payload)
	case msg.PayloadTypeFloat:
		pl, err = strconv.ParseFloat(m.Payload, 64)
	case msg.PayloadTypeInteger:
		pl, err = strconv.ParseInt(m.Payload, 10, 64)
	case msg.PayloadTypeString:
		pl = m.Payload
	case msg.PayloadTypeGeo:
		pl = m.Payload
	default:
		zap.L().Error("Unknown data type", zap.Any("sensorField", m))
	}

	if err != nil {
		zap.L().Error("Unable to convert the payload to actual PayloadType", zap.Error(err), zap.Any("sensorField", m))
	}

	// update payload
	cFv := sml.FieldValue{Value: pl, IsReceived: m.IsReceived, Timestamp: m.Timestamp}

	sf := &sml.SensorField{
		ID:          fmt.Sprintf("%s_%s_%s_%s", m.GatewayID, m.NodeID, m.SensorID, m.Field),
		ShortID:     m.Field,
		GatewayID:   m.GatewayID,
		NodeID:      m.NodeID,
		SensorID:    m.SensorID,
		LastSeen:    m.Timestamp,
		PayloadType: m.PayloadType,
		UnitID:      m.PayloadUnitID,
		Others:      m.Others,
		Payload:     cFv,
	}

	start := time.Now()
	f := []ml.Filter{
		{Key: "id", Operator: "eq", Value: sf.ID},
		// {Key: "gatewayId", Operator: "eq", Value: sf.GatewayID},
		// {Key: "nodeId", Operator: "eq", Value: sf.NodeID},
		// {Key: "sensorId", Operator: "eq", Value: sf.SensorID},
	}
	err = svc.STG.Upsert(ml.EntitySensorField, f, sf)
	if err != nil {
		zap.L().Error("Failed to update sensor field in to database", zap.Error(err), zap.Any("sensorField", sf))
	} else {
		zap.L().Debug("Inserted in to storage db", zap.String("timeTaken", time.Since(start).String()))
	}

	start = time.Now()
	err = svc.MTS.Write(sf)
	if err != nil {
		zap.L().Error("Failed to write into metrics database", zap.Error(err), zap.Any("sensorField", sf))
	} else {
		zap.L().Debug("Inserted in to metric db", zap.String("timeTaken", time.Since(start).String()))
	}
	return nil
}

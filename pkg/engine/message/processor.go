package messageprocessor

import (
	"fmt"
	"strconv"
	"time"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/mustafaturan/bus"
	"github.com/mycontroller-org/mycontroller-v2/pkg/mcbus"
	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	msg "github.com/mycontroller-org/mycontroller-v2/pkg/model/message"
	srv "github.com/mycontroller-org/mycontroller-v2/pkg/service"
	ut "github.com/mycontroller-org/mycontroller-v2/pkg/util"
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
	srv.BUS.RegisterHandler(mcbus.TopicMsgFromGW, &bus.Handler{
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
		// invalid message
		return
	}
	//zap.L().Debug("Message", zap.Any("message", m))

	// store it into database
	//srv.STG.Upsert(ml.EntityGateway,)
	// srv.STG.Upsert(ml.EntityNode, nil, nil)

	switch m.Command {
	case msg.CommandNode:
		n := &ml.Node{
			ID:        fmt.Sprintf("%s_%s", m.GatewayID, m.NodeID),
			GatewayID: m.GatewayID,
			ShortID:   m.NodeID,
			LastSeen:  m.Timestamp,
		}
		nodeData(n, m)
	case msg.CommandSensor:
		s := &ml.Sensor{
			ID:        fmt.Sprintf("%s_%s_%s", m.GatewayID, m.NodeID, m.SensorID),
			GatewayID: m.GatewayID,
			NodeID:    m.NodeID,
			ShortID:   m.SensorID,
			LastSeen:  m.Timestamp,
		}
		sensorData(s, m)
	case msg.CommandSet:
		sf := &ml.SensorField{
			ID:             fmt.Sprintf("%s_%s_%s_%s", m.GatewayID, m.NodeID, m.SensorID, m.Field),
			ShortID:        m.Field,
			GatewayID:      m.GatewayID,
			NodeID:         m.NodeID,
			SensorID:       m.SensorID,
			LastSeen:       m.Timestamp,
			PayloadType:    m.PayloadType,
			UnitID:         m.PayloadUnitID,
			ProviderConfig: m.Others,
		}
		sensorFieldData(sf, m)
	case msg.CommandRequest:
	case msg.CommandStream:
	case msg.CommandNone:
	}
	zap.L().Debug("Message process completed", zap.String("timeTaken", time.Since(m.Timestamp).String()), zap.Any("message", m))
}

func nodeData(n *ml.Node, m *msg.Message) error {
	f := []ml.Filter{
		{Key: "id", Operator: "eq", Value: n.ID},
	}
	node := &ml.Node{}
	err := srv.STG.FindOne(ml.EntityNode, f, node)
	if err != nil {
		node = n
		err = srv.STG.Insert(ml.EntityNode, node)
		if err != nil {
			zap.L().Error("Unable to insert the node in to database", zap.Error(err), zap.Any("node", node))
		}
	}
	if node.Others == nil {
		node.Others = map[string]interface{}{}
	}
	switch m.SubCommand {
	case msg.KeySubCmdName:
		node.Name = m.Payload
	case msg.KeySubCmdBatteryLevel:
		// update battery level
		bl, err := strconv.ParseFloat(m.Payload, 64)
		if err != nil {
			zap.L().Error("Unable to parse batter level", zap.Error(err))
		}
		node.Others[m.Field] = bl
		// send it to metric store
	case msg.KeySubCmdDiscover, msg.KeySubCmdHeartbeat, msg.KeySubCmdPing:
		// TODO:
	default:
		node.Others[m.Field] = m.Payload
	}
	if len(m.Others) > 0 {
		ut.JoinMap(node.Others, m.Others)
	}
	err = srv.STG.Upsert(ml.EntityNode, f, node)
	if err != nil {
		zap.L().Error("Unable to update the node in to database", zap.Error(err), zap.Any("node", node))
	}
	return nil
}

func sensorData(s *ml.Sensor, m *msg.Message) error {
	f := []ml.Filter{
		{Key: "id", Operator: "eq", Value: s.ID},
	}
	sensor := &ml.Sensor{}
	err := srv.STG.FindOne(ml.EntitySensor, f, sensor)
	if err != nil {
		sensor = s
		err = srv.STG.Insert(ml.EntitySensor, sensor)
		if err != nil {
			zap.L().Error("Unable to insert the sensor in to database", zap.Error(err), zap.Any("sensor", sensor))
		}
	}
	if sensor.ProviderConfig == nil {
		sensor.ProviderConfig = map[string]interface{}{}
	}

	// update sensor name
	if m.SubCommand == msg.KeyName {
		sensor.Name = m.Payload
	}
	if len(m.Others) > 0 {
		ut.JoinMap(sensor.ProviderConfig, m.Others)
	}
	err = srv.STG.Upsert(ml.EntitySensor, f, sensor)
	if err != nil {
		zap.L().Error("Unable to update the sensor in to database", zap.Error(err), zap.Any("sensor", sensor))
	}
	return nil
}

func sensorFieldData(sf *ml.SensorField, m *msg.Message) error {
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
	sf.Payload = ml.FieldValue{Value: pl}

	start := time.Now()
	f := []ml.Filter{
		{Key: "id", Operator: "eq", Value: sf.ID},
		// {Key: "gatewayId", Operator: "eq", Value: sf.GatewayID},
		// {Key: "nodeId", Operator: "eq", Value: sf.NodeID},
		// {Key: "sensorId", Operator: "eq", Value: sf.SensorID},
	}
	err = srv.STG.Upsert(ml.EntitySensorField, f, sf)
	if err != nil {
		zap.L().Error("Failed to update sensor field in to database", zap.Error(err), zap.Any("sensorField", sf))
	} else {
		zap.L().Debug("Inserted in to storage db", zap.String("timeTaken", time.Since(start).String()))
	}

	start = time.Now()
	err = srv.MTS.Write(sf)
	if err != nil {
		zap.L().Error("Failed to write into metrics database", zap.Error(err), zap.Any("sensorField", sf))
	} else {
		zap.L().Debug("Inserted in to metric db", zap.String("timeTaken", time.Since(start).String()))
	}

	return nil
}

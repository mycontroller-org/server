package messageprocessor

import (
	"fmt"
	"strconv"
	"time"

	q "github.com/jaegertracing/jaeger/pkg/queue"
	"github.com/mustafaturan/bus"
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
	srv.BUS.RegisterHandler("primary_engine", &bus.Handler{
		Matcher: srv.TopicGatewayMessageReceive,
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
	//zap.L().Debug("Received an item", zap.Any("item", e))
	m := e.Data.(*msg.Message)
	//zap.L().Debug("Message", zap.Any("message", m))

	// store it into database
	//srv.STG.Upsert(ml.EntityGateway,)
	// srv.STG.Upsert(ml.EntityNode, nil, nil)

	for _, fl := range m.Fields {
		switch fl.Command {
		case msg.CommandNode:
			n := &ml.Node{
				ID:        fmt.Sprintf("%s_%s", m.GatewayID, m.NodeID),
				GatewayID: m.GatewayID,
				ShortID:   m.NodeID,
				LastSeen:  m.Timestamp,
			}
			nodeData(n, &fl)
		case msg.CommandSensor:
			s := &ml.Sensor{
				ID:        fmt.Sprintf("%s_%s_%s", m.GatewayID, m.NodeID, m.SensorID),
				GatewayID: m.GatewayID,
				NodeID:    m.NodeID,
				ShortID:   m.SensorID,
				LastSeen:  m.Timestamp,
			}
			sensorData(s, &fl)
		case msg.CommandSet:
			sf := &ml.SensorField{
				ID:             fmt.Sprintf("%s_%s_%s_%s", m.GatewayID, m.NodeID, m.SensorID, fl.Key),
				ShortID:        fl.Key,
				GatewayID:      m.GatewayID,
				NodeID:         m.NodeID,
				SensorID:       m.SensorID,
				LastSeen:       m.Timestamp,
				DataType:       fl.DataType,
				UnitID:         fl.UnitID,
				ProviderConfig: fl.Others,
			}
			sensorFieldData(sf, &fl)
		case msg.CommandRequest:
		case msg.CommandStream:
		case msg.CommandNone:
		}
	}
	zap.L().Debug("Message process completed", zap.String("timeTaken", time.Since(m.Timestamp).String()), zap.Any("message", m))
}

func nodeData(n *ml.Node, fl *msg.Field) error {
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
	switch fl.Key {
	case msg.KeyCmdNodeName:
		node.Name = fl.Payload
	case msg.KeyCmdNodeBatteryLevel:
		// update battery level
		bl, err := strconv.ParseFloat(fl.Payload, 64)
		if err != nil {
			zap.L().Error("Unable to parse batter level", zap.Error(err))
		}
		node.Others[fl.Key] = bl
		// send it to metric store
	case msg.KeyCmdNodeDiscover, msg.KeyCmdNodeHeartbeat, msg.KeyCmdNodePing:
		// TODO:
	default:
		node.Others[fl.Key] = fl.Payload
	}
	if len(fl.Others) > 0 {
		ut.JoinMap(node.Others, fl.Others)
	}
	err = srv.STG.Upsert(ml.EntityNode, f, node)
	if err != nil {
		zap.L().Error("Unable to update the node in to database", zap.Error(err), zap.Any("node", node))
	}
	return nil
}

func sensorData(s *ml.Sensor, fl *msg.Field) error {
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
	if fl.Key == msg.KeyName {
		sensor.Name = fl.Payload
	}
	if len(fl.Others) > 0 {
		ut.JoinMap(sensor.ProviderConfig, fl.Others)
	}
	err = srv.STG.Upsert(ml.EntitySensor, f, sensor)
	if err != nil {
		zap.L().Error("Unable to update the sensor in to database", zap.Error(err), zap.Any("sensor", sensor))
	}
	return nil
}

func sensorFieldData(sf *ml.SensorField, fl *msg.Field) error {
	var err error
	var pl interface{}
	// convert payload to actual type
	switch fl.DataType {
	case msg.DataTypeBoolean:
		pl, err = strconv.ParseBool(fl.Payload)
	case msg.DataTypeFloat:
		pl, err = strconv.ParseFloat(fl.Payload, 64)
	case msg.DataTypeInteger:
		pl, err = strconv.ParseInt(fl.Payload, 10, 64)
	case msg.DataTypeString:
		pl = fl.Payload
	case msg.DataTypeGeo:
		pl = fl.Payload
	default:
		zap.L().Error("Unknown data type", zap.Any("sensorField", fl))
	}

	if err != nil {
		zap.L().Error("Unable to convert the payload to actual dataType", zap.Error(err), zap.Any("sensorField", fl))
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

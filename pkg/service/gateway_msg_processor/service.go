package gatewaymessageprocessor

import (
	"context"
	"errors"
	"fmt"
	"time"

	entityAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	eventTY "github.com/mycontroller-org/server/v2/pkg/types/event"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	"github.com/mycontroller-org/server/v2/pkg/types/topic"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/server/v2/pkg/utils/javascript"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	queueUtils "github.com/mycontroller-org/server/v2/pkg/utils/queue"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

const (
	unknownName = "unknown"
)

const (
	defaultQueueSize = int(1000)
	defaultWorkers   = int(1)
)

type MessageProcessor struct {
	ctx         context.Context
	logger      *zap.Logger
	queueName   string
	api         *entityAPI.API
	bus         busTY.Plugin
	metric      metricTY.Plugin
	eventsQueue *queueUtils.QueueSpec
}

func New(ctx context.Context, queueName string) (serviceTY.Service, error) {
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	api, err := entityAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	metric, err := metricTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	svc := &MessageProcessor{
		ctx:    ctx,
		logger: logger.Named("message_processor"),
		api:    api,
		bus:    bus,
		metric: metric,
	}

	svc.eventsQueue = &queueUtils.QueueSpec{
		Topic:          topic.TopicPostMessageToProcessor,
		Queue:          queueUtils.New(svc.logger, "message_processor", defaultQueueSize, svc.processMessage, defaultWorkers),
		SubscriptionId: -1,
	}

	return svc, nil
}

func (svc *MessageProcessor) Name() string {
	return "message_processor_service"
}

// Start message process engine
func (svc *MessageProcessor) Start() error {
	// on message receive add it in to our local queue
	if svc.queueName != "" {
		id, err := svc.bus.QueueSubscribe(svc.eventsQueue.Topic, svc.queueName, svc.onMessageReceive)
		if err != nil {
			return err
		}
		svc.eventsQueue.SubscriptionId = id
	} else {
		id, err := svc.bus.Subscribe(svc.eventsQueue.Topic, svc.onMessageReceive)
		if err != nil {
			return err
		}
		svc.eventsQueue.SubscriptionId = id
	}
	return nil
}

func (svc *MessageProcessor) onMessageReceive(busData *busTY.BusData) {
	msg := &msgTY.Message{}
	err := busData.LoadData(msg)
	if err != nil {
		svc.logger.Warn("Failed to convert to target type", zap.Error(err), zap.Any("busData", busData))
		return
	}

	if msg.GatewayID == "" {
		svc.logger.Warn("received an empty message", zap.Any("busData", busData))
		return
	}
	svc.logger.Debug("Message added into processing queue", zap.Any("message", msg))
	status := svc.eventsQueue.Produce(msg)
	if !status {
		svc.logger.Warn("Failed to store the message into queue", zap.Any("message", msg))
	}
}

// Close message process engine
func (svc *MessageProcessor) Close() error {
	svc.eventsQueue.Close()
	return nil
}

// processMessage from the queue
func (svc *MessageProcessor) processMessage(item interface{}) {
	msg := item.(*msgTY.Message)
	svc.logger.Debug("Starting Message Processing", zap.Any("message", msg))

	switch {
	case msg.SourceID != "":
		switch msg.Type {
		case msgTY.TypeSet: // set fields
			err := svc.setFieldData(msg)
			if err != nil {
				svc.logger.Error("error on field data set", zap.Error(err))
			}
			// update last seen
			svc.updateSourceLastSeen(msg.GatewayID, msg.NodeID, msg.SourceID, msg.Timestamp)
			svc.updateNodeLastSeen(msg.GatewayID, msg.NodeID, msg.Timestamp)

		case msgTY.TypeRequest: // request fields
			err := svc.requestFieldData(msg)
			if err != nil {
				svc.logger.Error("error on field data request", zap.Error(err))
			}

		case msgTY.TypePresentation: // update source data, like name or other details
			err := svc.updateSourceDetail(msg)
			if err != nil {
				svc.logger.Error("error on source data update", zap.Error(err))
			}
			// update last seen
			svc.updateSourceLastSeen(msg.GatewayID, msg.NodeID, msg.SourceID, msg.Timestamp)
			svc.updateNodeLastSeen(msg.GatewayID, msg.NodeID, msg.Timestamp)

		default:
			svc.logger.Warn("message type not implemented for source", zap.String("type", msg.Type), zap.Any("message", msg))
		}

	case msg.NodeID != "":
		switch msg.Type {
		case msgTY.TypeSet, msgTY.TypePresentation: // set node specific data, like battery level, rssi, etc
			err := svc.updateNodeData(msg)
			if err != nil {
				svc.logger.Error("error on node data update", zap.Error(err))
			}
			// node last seen managed in updateNodeData

		case msgTY.TypeRequest: // request node specific data

		case msgTY.TypeAction: // providers will take care of action type messages
			clonedMsg := msg.Clone() // clone the message
			svc.postMessage(clonedMsg)

		default:
			svc.logger.Warn("message type not implemented for node", zap.String("type", msg.Type), zap.Any("message", msg))
		}

	case msg.NodeID == "" && msg.Type == msgTY.TypeAction:
		clonedMsg := msg.Clone() // clone the message
		svc.postMessage(clonedMsg)

	default:
		svc.logger.Warn("this message not handled", zap.Any("message", msg))
	}

	svc.logger.Debug("message processed", zap.String("timeTaken", time.Since(msg.Timestamp).String()), zap.Any("message", msg))
}

// update node detail
func (svc *MessageProcessor) updateNodeData(msg *msgTY.Message) error {
	node, err := svc.api.Node().GetByGatewayAndNodeID(msg.GatewayID, msg.NodeID)
	if err != nil {
		if err == storageTY.ErrNoDocuments {
			node = &nodeTY.Node{
				GatewayID: msg.GatewayID,
				NodeID:    msg.NodeID,
				Name:      unknownName,
			}
		} else {
			svc.logger.Error("error on getting node data", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
			return err
		}
	}

	// init labels and others
	node.Labels = node.Labels.Init()
	node.Others = node.Others.Init()

	// update last seen
	node.LastSeen = msg.Timestamp

	// update node status
	if node.State.Status != types.StatusUp {
		node.State = types.State{
			Status: types.StatusUp,
			Since:  msg.Timestamp,
		}
	}

	for _, d := range msg.Payloads {
		// update labels
		node.Labels.CopyFrom(d.Labels)

		switch d.Key { // set node name
		case types.FieldName:
			if !node.Labels.GetIgnoreBool(types.LabelName) {
				node.Name = d.Value.String()
			}

		case types.FieldBatteryLevel: // set battery level
			// update battery level
			batteryLevel := converterUtils.ToFloat(d.Value.String())
			node.Others.Set(d.Key, batteryLevel, node.Labels)
			err = svc.writeNodeMetric(node, metricTY.MetricTypeGaugeFloat, types.FieldBatteryLevel, batteryLevel)
			if err != nil {
				svc.logger.Error("error on writing metric data", zap.Error(err))
			}

		default:
			if d.Key != types.FieldNone {
				node.Others.Set(d.Key, d.Value.String(), node.Labels)
				// TODO: Do we need to report to metric strore?
			}

		}

		// update labels and Others
		node.Labels.CopyFrom(d.Labels)
		node.Others.CopyFrom(d.Others, node.Labels)
	}

	// save node data and publish events
	err = svc.api.Node().Save(node, true)
	if err != nil {
		svc.logger.Error("unable to update save the node data", zap.Error(err), zap.Any("node", node))
		return err
	}

	return nil
}

func (svc *MessageProcessor) updateSourceDetail(msg *msgTY.Message) error {
	source, err := svc.api.Source().GetByIDs(msg.GatewayID, msg.NodeID, msg.SourceID)
	if err != nil {
		if err == storageTY.ErrNoDocuments {
			source = &sourceTY.Source{
				GatewayID: msg.GatewayID,
				NodeID:    msg.NodeID,
				SourceID:  msg.SourceID,
				Name:      unknownName,
			}
		} else {
			svc.logger.Error("error on getting source data", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.String("sourceId", msg.SourceID), zap.Error(err))
			return err
		}
	}

	// update last seen
	source.LastSeen = msg.Timestamp

	// init labels and others
	source.Labels = source.Labels.Init()
	source.Others = source.Others.Init()

	for _, payload := range msg.Payloads {
		switch payload.Key {
		case types.FieldName: // set name
			if !source.Labels.GetIgnoreBool(types.LabelName) {
				source.Name = payload.Value.String()
			}

		default: // set other variables
			if payload.Key != types.FieldNone {
				source.Others.Set(payload.Key, payload.Value.String(), source.Labels)
				// TODO: Do we need to report to metric strore?
			}
		}

		// update labels and Others
		source.Labels.CopyFrom(payload.Labels)
		source.Others.CopyFrom(payload.Others, source.Labels)
	}

	err = svc.api.Source().Save(source)
	if err != nil {
		svc.logger.Error("unable to update the source in to database", zap.Error(err), zap.Any("source", source))
		return err
	}
	// post field data to event listeners
	busUtils.PostEvent(svc.logger, svc.bus, topic.TopicEventSource, eventTY.TypeUpdated, types.EntitySource, source)
	return nil
}

func (svc *MessageProcessor) setFieldData(msg *msgTY.Message) error {
	for _, payload := range msg.Payloads {
		field, err := svc.api.Field().GetByIDs(msg.GatewayID, msg.NodeID, msg.SourceID, payload.Key)
		if err != nil {
			if err == storageTY.ErrNoDocuments {
				field = &fieldTY.Field{
					GatewayID: msg.GatewayID,
					NodeID:    msg.NodeID,
					SourceID:  msg.SourceID,
					FieldID:   payload.Key,
					Name:      unknownName,
				}
			} else {
				svc.logger.Error("error on getting field data", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.String("sourceId", msg.SourceID), zap.String("fieldId", payload.Key), zap.Error(err))
				return err
			}
		}

		value := payload.Value.String()

		// if custom payload formatter supplied
		if formatter := field.Formatter.OnReceive; msg.IsReceived && formatter != "" {
			startTime := time.Now()

			scriptInput := map[string]interface{}{
				"value":         payload.Value.String(),
				"lastValue":     field.Current.Value,
				"previousValue": field.Previous.Value,
			}

			// script runs without timeout
			// TODO: include timeout
			responseValue, err := javascript.Execute(svc.logger, formatter, scriptInput, nil)
			if err != nil {
				svc.logger.Error("error on executing script", zap.Error(err), zap.Any("inputValue", payload.Value.String()), zap.String("gateway", field.GatewayID), zap.String("node", field.NodeID), zap.String("source", field.SourceID), zap.String("fieldID", field.FieldID), zap.String("script", formatter))
				return err
			}

			formattedValue := ""
			if responseValue == nil {
				svc.logger.Error("returned nil value", zap.String("formatter", formatter))
				return errors.New("formatter returned nil value")
			}

			if _sliceOfMap, ok := responseValue.([]interface{}); ok { // if the formatted response is slice of map
				for _, _mapData := range _sliceOfMap {
					_fieldMap, ok := _mapData.(map[string]interface{})
					if !ok { // return from here
						svc.logger.Error("error on converting supplied data", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.String("sourceId", msg.SourceID), zap.String("fieldId", field.FieldID), zap.Any("input", _mapData))
						return errors.New("supplied input not in map[string]interface{} type")
					}
					err = svc.updateFieldWithFormattedData(msg, field, _fieldMap)
					if err != nil {
						return err
					}
				}
				return nil
			} else if _fieldMap, ok := responseValue.(map[string]interface{}); ok { // if the formatted response is map
				return svc.updateFieldWithFormattedData(msg, field, _fieldMap)
			} else { // if non of the above
				formattedValue = converterUtils.ToString(responseValue)
			}

			// update the formatted value
			svc.logger.Debug("formatting done", zap.Any("oldValue", payload.Value.String()), zap.String("newValue", formattedValue), zap.String("timeTaken", time.Since(startTime).String()))
			// update formatted value into value
			value = formattedValue
		}

		err = svc.updateFieldData(field, payload.Key, payload.Key, payload.MetricType, payload.Unit, payload.Labels, payload.Others, value, msg)
		if err != nil {
			svc.logger.Error("error on updating field data", zap.Error(err), zap.String("gateway", msg.GatewayID), zap.String("node", msg.NodeID), zap.String("source", msg.SourceID), zap.String("field", payload.Key))
		}
	}
	return nil
}

// updates field data with received map data
func (svc *MessageProcessor) updateFieldWithFormattedData(msg *msgTY.Message, field *fieldTY.Field, mapData map[string]interface{}) error {
	if _, found := mapData["value"]; !found {
		svc.logger.Error("value field not updated", zap.Any("received", mapData), zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.String("sourceId", msg.SourceID), zap.String("fieldId", field.FieldID))
		return errors.New("formatter returned nil value")
	}
	_field := field.Clone()
	// reset labels and others
	_field.Labels = cmap.CustomStringMap{}
	_field.Others = cmap.CustomMap{}

	for key, _value := range mapData {
		switch key {
		case "fieldId":
			_field.FieldID = converterUtils.ToString(_value)
		case "sourceId":
			_field.SourceID = converterUtils.ToString(_value)
		case "value":
			_field.Current.Value = _value
		case "name":
			_field.Name = converterUtils.ToString(_value)
		case "unit":
			_field.Unit = converterUtils.ToString(_value)
		case "metricType":
			_field.MetricType = converterUtils.ToString(_value)
		case "labels":
			if _labels, ok := _value.(map[string]string); ok {
				_field.Labels.CopyFrom(_labels)
			} else if _labels, ok := _value.(map[string]interface{}); ok {
				for l_key, l_value := range _labels {
					_field.Labels.Set(l_key, converterUtils.ToString(l_value))
				}
			}
		case "others":
			if _others, ok := _value.(map[string]interface{}); ok {
				_field.Others.CopyFrom(_others, nil)
			}
		}
	}

	// get the actual field
	actualField, err := svc.api.Field().GetByIDs(msg.GatewayID, msg.NodeID, _field.SourceID, _field.FieldID)
	if err != nil {
		if err != storageTY.ErrNoDocuments {
			svc.logger.Error("error on getting field data", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.String("sourceId", _field.SourceID), zap.String("fieldId", _field.FieldID), zap.Error(err))
			return err
		}
		actualField = &fieldTY.Field{
			GatewayID: msg.GatewayID,
			NodeID:    msg.NodeID,
			SourceID:  _field.SourceID,
			FieldID:   _field.FieldID,
			Name:      unknownName,
		}
	}

	err = svc.updateFieldData(actualField, _field.FieldID, _field.Name, _field.MetricType, _field.Unit, _field.Labels, _field.Others, _field.Current.Value, msg)
	if err != nil {
		svc.logger.Error("error on updating field data", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.String("sourceId", _field.SourceID), zap.String("fieldId", _field.FieldID), zap.Error(err))
		return err
	}
	return nil
}

func (svc *MessageProcessor) updateFieldData(
	field *fieldTY.Field, fieldId, name, metricType, unit string, labels cmap.CustomStringMap,
	others cmap.CustomMap, value interface{}, msg *msgTY.Message) error {

	// if metricType is empty update as none
	if metricType == "" {
		metricType = metricTY.MetricTypeNone
	}

	if field == nil {
		updateField, err := svc.api.Field().GetByIDs(msg.GatewayID, msg.NodeID, msg.SourceID, fieldId)
		if err != nil {
			if err != storageTY.ErrNoDocuments {
				svc.logger.Error("error on getting field data", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.String("sourceId", msg.SourceID), zap.String("fieldId", fieldId), zap.Error(err))
				return err
			}
			field = &fieldTY.Field{
				GatewayID: msg.GatewayID,
				NodeID:    msg.NodeID,
				SourceID:  msg.SourceID,
				FieldID:   fieldId,
				Name:      unknownName,
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
	if !field.Labels.GetIgnoreBool(types.LabelName) {
		field.Name = name
	}

	// update type
	if !field.Labels.GetIgnoreBool(types.LabelMetricType) {
		field.MetricType = metricType
	}
	// update unit
	if !field.Labels.GetIgnoreBool(types.LabelUnit) {
		field.Unit = unit
	}

	svc.logger.Debug("field", zap.Any("field", field))
	// update labels and others
	field.Labels.CopyFrom(labels)               // copy labels
	field.Others.CopyFrom(others, field.Labels) // copy other fields

	// convert value to specified metric type
	// convert payload to actual type
	var convertedValue interface{}
	switch field.MetricType {

	case metricTY.MetricTypeBinary:
		convertedValue = converterUtils.ToBool(value)

	case metricTY.MetricTypeGaugeFloat:
		convertedValue = converterUtils.ToFloat(value)

	case metricTY.MetricTypeGauge, metricTY.MetricTypeCounter:
		convertedValue = converterUtils.ToInteger(value)

	case metricTY.MetricTypeNone:
		convertedValue = value

	case metricTY.MetricTypeString:
		convertedValue = converterUtils.ToString(value)

	case metricTY.MetricTypeGEO: // Implement geo
		convertedValue = value

	default:
		svc.logger.Error("unknown data type on a field", zap.Any("message", msg))
		return fmt.Errorf("unknown metricType: %s", field.MetricType)
	}

	// update shift old payload and update current payload
	field.Previous = field.Current
	field.Current = fieldTY.Payload{Value: convertedValue, IsReceived: msg.IsReceived, Timestamp: msg.Timestamp}

	// update no change since
	oldValue := convertor.ToString(field.Previous.Value)
	newValue := convertor.ToString(field.Current.Value)
	if oldValue != newValue {
		field.NoChangeSince = msg.Timestamp
	}

	startTime := time.Now()
	err := svc.api.Field().Save(field, false)
	if err != nil {
		svc.logger.Error("failed to update field in to database", zap.Error(err), zap.Any("field", field))
	} else {
		svc.logger.Debug("inserted in to storage db", zap.String("timeTaken", time.Since(startTime).String()))
	}

	// post field data to event listeners
	busUtils.PostEvent(svc.logger, svc.bus, topic.TopicEventField, eventTY.TypeUpdated, types.EntityField, field)

	updateMetric := true
	if field.MetricType == metricTY.MetricTypeNone {
		updateMetric = false
	}
	// for binary do not update duplicate values
	if field.MetricType == metricTY.MetricTypeBinary {
		updateMetric = field.Current.Timestamp.Equal(field.NoChangeSince)
	}
	if updateMetric {
		err = svc.writeFieldMetric(field)
		if err != nil {
			return err
		}
	} else {
		svc.logger.Debug("skipped metric update", zap.Any("field", field))
	}
	return nil
}

func (svc *MessageProcessor) requestFieldData(msg *msgTY.Message) error {
	payloads := make([]msgTY.Payload, 0)
	for _, payload := range msg.Payloads {
		field, err := svc.api.Field().GetByIDs(msg.GatewayID, msg.NodeID, msg.SourceID, payload.Key)
		if err != nil {
			// TODO: check availability error message from storage
			continue
		}

		if field.Current.Value != nil {
			payload.SetValue(fmt.Sprintf("%v", field.Current.Value)) // update payload
			if payload.Value.String() != "" {                        // if the value is not empty update it
				payload.Labels = field.Labels.Clone()
				clonedData := payload.Clone() // clone the message
				payloads = append(payloads, clonedData)
			}
		}
		// post field data to event listeners
		// NOTE: if the entry not available in database, request will be dropped
		busUtils.PostEvent(svc.logger, svc.bus, topic.TopicEventField, eventTY.TypeRequested, types.EntityField, field)
	}

	if len(payloads) > 0 {
		clonedMsg := msg.Clone()         // clone the message
		clonedMsg.Timestamp = time.Now() // set current timestamp
		clonedMsg.Payloads = payloads    // update payload
		clonedMsg.Type = msgTY.TypeSet   // change type to set
		clonedMsg.IsSleepNode = false    // response immediately, do not check "is a sleeping node"
		svc.postMessage(clonedMsg)
	} else {
		svc.logger.Debug("no data found for this request", zap.Any("message", msg))
	}
	return nil
}

// topic to send message to provider gateway
func (svc *MessageProcessor) postMessage(msg *msgTY.Message) {
	if msg.IsAck {
		return // do not respond for ack message
	}
	// register the node, if not available
	// example: MySensors auto node id generation, register it on firmware request
	if msg.NodeID != "" {
		node, _ := svc.api.Node().GetByGatewayAndNodeID(msg.GatewayID, msg.NodeID)
		if node == nil {
			node = &nodeTY.Node{
				GatewayID: msg.GatewayID,
				NodeID:    msg.NodeID,
				Name:      "unknown",
			}
			// save node data
			err := svc.api.Node().Save(node, true)
			if err != nil {
				svc.logger.Error("unable to update save the node data", zap.Error(err), zap.Any("node", node))
				return
			}
		}
	}
	topic := fmt.Sprintf("%s.%s", topic.TopicPostMessageToProvider, msg.GatewayID)
	msg.IsReceived = false
	// include node labels
	node, err := svc.api.Node().GetByGatewayAndNodeID(msg.GatewayID, msg.NodeID)
	if err != nil {
		svc.logger.Debug("error on getting node details", zap.String("gatewayId", msg.GatewayID), zap.String("nodeId", msg.NodeID), zap.Error(err))
	} else {
		msg.Labels = msg.Labels.Init()
		msg.Labels.CopyFrom(node.Labels)
	}
	svc.logger.Debug("posting message", zap.String("topic", topic), zap.Any("message", msg))
	err = svc.bus.Publish(topic, msg)
	if err != nil {
		svc.logger.Error("error on posting message", zap.String("topic", topic), zap.Any("message", msg), zap.Error(err))
	}
}

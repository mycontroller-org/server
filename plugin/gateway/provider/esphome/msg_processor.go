package esphome

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/model"
	msgML "github.com/mycontroller-org/server/v2/pkg/model/message"
	"github.com/mycontroller-org/server/v2/pkg/service/mcbus"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	colorUtils "github.com/mycontroller-org/server/v2/pkg/utils/color"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/server/v2/plugin/database/metrics"
	esphomeAPI "github.com/mycontroller-org/esphome_api/pkg/api"
	"go.uber.org/zap"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Post sends a command to esphome node
func (p *Provider) Post(message *msgML.Message) error {
	if message == nil || len(message.Payloads) == 0 || message.NodeID == "" {
		zap.L().Error("invalid message received", zap.String("gatewayId", p.GatewayConfig.ID), zap.Any("message", message))
		return errors.New("invalid message")
	}

	espNode := p.clientStore.Get(message.NodeID)
	if espNode == nil {
		return nil
	}

	if message.Type == msgML.TypeAction {
		return p.handleActions(message)
	}

	if message.SourceID == "" {
		zap.L().Error("invalid message received", zap.String("gatewayId", p.GatewayConfig.ID), zap.Any("message", message))
		return errors.New("sourceID not found")
	}

	entity := p.entityStore.GetBySourceID(message.NodeID, message.SourceID)
	if entity == nil {
		return nil
	}

	payload := message.Payloads[0]
	fieldID := strings.ToLower(payload.Key)

	fields := make(map[string]interface{})
	fields[FieldKey] = entity.Key
	fields[fmt.Sprintf("has_%s", fieldID)] = true

	var request protoreflect.ProtoMessage

	switch entity.Type {

	case EntityTypeCamera:
		request = &esphomeAPI.CameraImageRequest{}
		fields[fieldID] = convertor.ToBool(payload.Value)

	case EntityTypeClimate:
		request = &esphomeAPI.ClimateCommandRequest{}
		adjustValueToEsphomeNode(entity.Type, fieldID, payload.Value, fields)

	case EntityTypeCover:
		request = &esphomeAPI.CoverCommandRequest{}
		adjustValueToEsphomeNode(entity.Type, fieldID, payload.Value, fields)

	case EntityTypeFan:
		request = &esphomeAPI.FanCommandRequest{}
		adjustValueToEsphomeNode(entity.Type, fieldID, payload.Value, fields)

	case EntityTypeLight:
		request = &esphomeAPI.LightCommandRequest{}
		adjustValueToEsphomeNode(entity.Type, fieldID, payload.Value, fields)

	case EntityTypeSwitch:
		request = &esphomeAPI.SwitchCommandRequest{}
		fields[fieldID] = convertor.ToBool(payload.Value)

	default:
		return nil
	}

	if request != nil {
		zap.L().Debug("field populated", zap.String("gatewayId", p.GatewayConfig.ID), zap.Any("fields", fields), zap.Any("entity", entity))

		jsonBytes, err := json.Marshal(fields)
		if err != nil {
			return err
		}

		err = json.Unmarshal(jsonBytes, request)
		if err != nil {
			return err
		}
		err = espNode.Post(request)
		if err != nil {
			return err
		}
		// create a echo to MyController to update the actual component
		// this block will be used only, when the field not updated by the esphome node
		if entity.Type == EntityTypeLight && fieldID == FieldRGB {
			msgSource := getMessage(p.GatewayConfig.ID, message.NodeID, message.SourceID, msgML.TypeSet, time.Now())
			sourceData := msgML.NewPayload()
			sourceData.Key = fieldID
			sourceData.Value = payload.Value
			sourceData.MetricType = metrics.MetricTypeNone
			msgSource.Payloads = append(msgSource.Payloads, sourceData)
			err = mcbus.Publish(mcbus.GetTopicPostMessageToServer(), msgSource)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// ProcessReceived performs operation on raw message received from esphome node and returns multiple messages
func (p *Provider) ProcessReceived(rawMsg *msgML.RawMessage) ([]*msgML.Message, error) {
	if rawMsg == nil {
		return nil, nil
	}
	zap.L().Debug("processing a message", zap.String("gatewayId", p.GatewayConfig.ID), zap.Any("type", fmt.Sprintf("%T", rawMsg.Data)), zap.Any("rawMessage", rawMsg))
	nodeID := rawMsg.Others.GetString(NodeID)
	protoMsg := rawMsg.Data.(protoreflect.ProtoMessage)

	missingStateSet := false
	switch actualMsg := rawMsg.Data.(type) {
	case *esphomeAPI.SensorStateResponse:
		missingStateSet = actualMsg.MissingState

	case *esphomeAPI.BinarySensorStateResponse:
		missingStateSet = actualMsg.MissingState

	case *esphomeAPI.TextSensorStateResponse:
		missingStateSet = actualMsg.MissingState

	}
	if missingStateSet {
		return nil, nil
	}

	fields, err := toFieldsMap(protoMsg)
	if err != nil {
		return nil, err
	}

	zap.L().Debug("fields", zap.String("gatewayId", p.GatewayConfig.ID), zap.Any("fields", fields))
	delete(fields, "missing_state") // this field is not required

	switch protoMsg.(type) {

	case *esphomeAPI.ListEntitiesBinarySensorResponse:
		return p.getMessageEntitiesResponse(EntityTypeBinarySensor, nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.ListEntitiesCameraResponse:
		return p.getMessageEntitiesResponse(EntityTypeCamera, nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.ListEntitiesClimateResponse:
		return p.getMessageEntitiesResponse(EntityTypeClimate, nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.ListEntitiesCoverResponse:
		return p.getMessageEntitiesResponse(EntityTypeCover, nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.ListEntitiesFanResponse:
		return p.getMessageEntitiesResponse(EntityTypeFan, nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.ListEntitiesLightResponse:
		return p.getMessageEntitiesResponse(EntityTypeLight, nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.ListEntitiesSwitchResponse:
		return p.getMessageEntitiesResponse(EntityTypeSwitch, nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.ListEntitiesSensorResponse:
		return p.getMessageEntitiesResponse(EntityTypeSensor, nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.ListEntitiesTextSensorResponse:
		return p.getMessageEntitiesResponse(EntityTypeTextSensor, nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.LightStateResponse,
		*esphomeAPI.BinarySensorStateResponse,
		*esphomeAPI.SwitchStateResponse,
		*esphomeAPI.FanStateResponse:
		if _, found := fields[FieldState]; !found {
			fields[FieldState] = false
		}
		return p.getStateResponse(nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.TextSensorStateResponse:
		if _, found := fields[FieldState]; !found {
			fields[FieldState] = ""
		}
		return p.getStateResponse(nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.CameraImageResponse:
		delete(fields, "done") // remove done field, not required
		return p.getStateResponse(nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.SensorStateResponse:
		return p.getStateResponse(nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.ClimateStateResponse,
		*esphomeAPI.CoverStateResponse:
		delete(fields, "legacy_state")
		return p.getStateResponse(nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.PingRequest:
		return p.getActionMessage(ActionPingRequest, nodeID, rawMsg.Timestamp, fields)

	case *esphomeAPI.GetTimeRequest:
		return p.getActionMessage(ActionTimeRequest, nodeID, rawMsg.Timestamp, fields)

	default:
		zap.L().Debug("unknown message received", zap.Any("type", fmt.Sprintf("%T", protoMsg)),
			zap.Any("message", protoMsg), zap.String("gatewayId", p.GatewayConfig.ID), zap.String("nodeId", nodeID))
	}

	return nil, nil
}

// toFieldsMap converts the proto.Message to map instance
// Array and slice values encode as JSON arrays, except that []byte encodes as a base64-encoded string,
// and a nil slice encodes as the null JSON object.
// https://golang.org/pkg/encoding/json/#Marshal
func toFieldsMap(espMsg interface{}) (map[string]interface{}, error) {
	bytes, err := json.Marshal(espMsg)
	if err != nil {
		return nil, err
	}
	fields := make(map[string]interface{})
	err = json.Unmarshal(bytes, &fields)
	if err != nil {
		return nil, err
	}
	return fields, nil
}

// some of the field has to be renamed to keep align with state response name
var fieldReMap = map[string]string{
	"white_value": "white",
	"oscillation": "oscillating",
}

func isField(fieldID string) (string, bool) {
	if strings.HasPrefix(fieldID, "supports") {
		field := strings.TrimPrefix(fieldID, "supports_")
		if newField, found := fieldReMap[field]; found {
			return newField, true
		}
		return field, true
	}
	return "", false
}

func adjustValueToMyController(entityType, fieldID string, value interface{}) string {
	switch entityType {
	case EntityTypeCamera:
		// if it is a camera image encode the bytes to base64
		if fieldID == FieldData {
			if bytes, ok := value.(string); ok {
				return fmt.Sprintf("%s%s", CameraImagePrefix, bytes)
			}
		}

	case EntityTypeLight:
		if fieldID == FieldBrightness {
			floatValue := convertor.ToFloat(value) * 100.0
			return convertor.ToString(floatValue)
		}

	}
	return convertor.ToString(value)
}

func adjustValueToEsphomeNode(entityType, fieldID string, value interface{}, fields map[string]interface{}) {
	switch entityType {

	case EntityTypeClimate:
		switch fieldID {
		case FieldMode, FieldFanMode, FieldSwingMode:
			fields[fieldID] = int32(convertor.ToInteger(value))
			return
		case FieldAway:
			fields[fieldID] = convertor.ToBool(value)
			return
		default:
			fields[fieldID] = float32(convertor.ToFloat(value))
			return
		}

	case EntityTypeCover:
		fields[fieldID] = convertor.ToBool(value)
		return

	case EntityTypeFan:
		switch fieldID {
		case FieldState, FieldOscillating:
			fields[fieldID] = convertor.ToBool(value)
			return
		default:
			fields[fieldID] = int32(convertor.ToInteger(value))
			return
		}

	case EntityTypeLight:
		switch fieldID {
		case FieldState:
			fields[fieldID] = convertor.ToBool(value)
			return
		case FieldBrightness:
			fields[fieldID] = float32(convertor.ToFloat(value) / 100.0)
			return
		case FieldWhite:
			fields[fieldID] = float32(convertor.ToFloat(value) / 100.0)
			return
		case FieldRGB:
			// convert RGB to r,g,b and send it
			red := float32(0)
			green := float32(0)
			blue := float32(0)

			stringValue := convertor.ToString(value)
			if strings.HasPrefix(stringValue, "#") && len(stringValue) == 7 {
				// example: #ad4e00
				red = colorUtils.HexToFloat32(stringValue[1:3])
				green = colorUtils.HexToFloat32(stringValue[3:5])
				blue = colorUtils.HexToFloat32(stringValue[5:])
			} else { // example: 254 (hue)
				// keeps saturation 99 all the time
				red, green, blue = colorUtils.ToRGB(convertor.ToFloat(value), 99, 0.01)
			}

			fields[FieldRed] = red / 255
			fields[FieldGreen] = green / 255
			fields[FieldBlue] = blue / 255
			return
		case FieldRed, FieldGreen, FieldBlue:
			fields["has_rgb"] = true
			fields[fieldID] = float32(convertor.ToFloat(value)) / 255.0
			return

		default:
			fields[fieldID] = float32(convertor.ToFloat(value))
			return
		}

	}
	fields[fieldID] = convertor.ToString(value)
}

// getMessageEntitiesResponse returns the entities as multiple messages
func (p *Provider) getMessageEntitiesResponse(entityType, nodeID string, timestamp time.Time, fields map[string]interface{}) ([]*msgML.Message, error) {
	zap.L().Debug("fields", zap.String("gatewayId", p.GatewayConfig.ID), zap.String("nodeId", nodeID), zap.Any("fields", fields))
	objectID, found := fields[FieldObjectID]
	if !found {
		return nil, errors.New("object id not found")
	}
	sourceID := convertor.ToString(objectID)

	keyRaw, found := fields[FieldKey]
	if !found {
		return nil, errors.New("key not found")
	}
	key := uint32(convertor.ToInteger(keyRaw))
	keyString := convertor.ToString(key)
	name := convertor.ToString(utils.GetMapValue(fields, FieldName, objectID))

	// delete common fields
	delete(fields, FieldKey)
	delete(fields, FieldName)
	delete(fields, FieldObjectID)
	delete(fields, FieldUniqueID)

	messages := make([]*msgML.Message, 0)

	msgSource := getMessage(p.GatewayConfig.ID, nodeID, sourceID, msgML.TypePresentation, timestamp)
	sourceData := msgML.NewPayload()
	sourceData.Key = model.FieldName
	sourceData.Value = name
	sourceData.Labels.Set(LabelKey, keyString)
	sourceData.Labels.Set(LabelType, entityType)

	msgFields := getMessage(p.GatewayConfig.ID, nodeID, sourceID, msgML.TypeSet, timestamp)

	unitOfMeasurement := ""
	deviceClass := ""
	// update presentation messages
	for field, value := range fields {
		zap.L().Debug("field", zap.String("gatewayId", p.GatewayConfig.ID), zap.String("field", field), zap.Any("value", value))
		if strings.HasPrefix(field, "legacy") { // ignore legacy fields
			continue
		} else if field == "unit_of_measurement" {
			unitOfMeasurement = convertor.ToString(value)
			continue
		} else if field == "device_class" {
			deviceClass = convertor.ToString(value)
			continue
		}

		fieldID, validField := isField(field)
		if validField {
			metricMap := getMetricData(entityType, fieldID)
			fieldData := msgML.NewPayload()
			fieldData.Key = fieldID
			fieldData.Value = ""
			fieldData.MetricType = metricMap.MetricType
			fieldData.Unit = metricMap.Unit
			fieldData.Labels.Set(model.LabelReadOnly, metricMap.ReadOnly)
			fieldData.Labels.Set(LabelKey, keyString)
			fieldData.Labels.Set(LabelType, entityType)
			msgFields.Payloads = append(msgFields.Payloads, fieldData)
		} else {
			sourceData.Others.Set(field, value, nil)
		}
	}

	// add it to local map
	entity := Entity{
		Key:         key,
		Type:        entityType,
		SourceID:    sourceID,
		SourceName:  name,
		DeviceClass: deviceClass,
		Unit:        unitOfMeasurement,
	}
	p.entityStore.AddEntity(nodeID, key, entity)

	if deviceClass != "" {
		sourceData.Labels.Set("device_class", deviceClass)
	}

	// fields with false value will not be in the fields map
	switch entityType {
	// case EntityTypeLight,
	// 	EntityTypeBinarySensor,
	// 	EntityTypeSwitch:
	// 	fieldState := msgML.NewData()
	// 	fieldState.Name = FieldState
	// 	fieldState.MetricType = metrics.MetricTypeBinary
	// 	fieldState.Labels.Set(LabelKey, keyString)
	// 	fieldState.Labels.Set(LabelType, entity.Type)
	// 	fieldState.Labels.Set(model.LabelReadOnly, "false")
	// 	if deviceClass != "" {
	// 		fieldState.Labels.Set("device_class", deviceClass)
	// 	}
	// 	msgFields.Payloads = append(msgFields.Payloads, fieldState)
	//
	// case EntityTypeSensor:
	// 	fieldState := msgML.NewData()
	// 	fieldState.Name = FieldState
	// 	fieldState.MetricType = metrics.MetricTypeGaugeFloat
	// 	fieldState.Unit = unitOfMeasurement
	// 	fieldState.Labels.Set(LabelKey, keyString)
	// 	fieldState.Labels.Set(LabelType, entity.Type)
	// 	fieldState.Labels.Set(model.LabelReadOnly, "true")
	// 	msgFields.Payloads = append(msgFields.Payloads, fieldState)
	//
	// case EntityTypeTextSensor:
	// 	fieldState := msgML.NewData()
	// 	fieldState.Name = FieldState
	// 	fieldState.MetricType = metrics.MetricTypeString
	// 	fieldState.Labels.Set(LabelKey, keyString)
	// 	fieldState.Labels.Set(LabelType, entity.Type)
	// 	fieldState.Labels.Set(model.LabelReadOnly, "true")
	// 	msgFields.Payloads = append(msgFields.Payloads, fieldState)

	// in camera, to enable or disable the stream, add 'stream' field manually
	case EntityTypeCamera:
		fieldStream := msgML.NewPayload()
		fieldStream.Key = FieldStream
		fieldStream.MetricType = metrics.MetricTypeNone
		fieldStream.Labels.Set(LabelKey, keyString)
		fieldStream.Labels.Set(LabelType, entity.Type)
		fieldStream.Labels.Set(model.LabelReadOnly, "false")
		msgFields.Payloads = append(msgFields.Payloads, fieldStream)

	}

	// include source data
	msgSource.Payloads = append(msgSource.Payloads, sourceData)

	messages = append(messages, msgSource)
	messages = append(messages, msgFields)

	zap.L().Debug("presentations", zap.String("gatewayId", p.GatewayConfig.ID), zap.Any("messages", messages))
	return messages, nil
}

// getStateResponse returns the status to multiple messages
func (p *Provider) getStateResponse(nodeID string, timestamp time.Time, fields map[string]interface{}) ([]*msgML.Message, error) {
	key, found := fields[FieldKey]
	if !found {
		zap.L().Info("key not found", zap.String("gatewayId", p.GatewayConfig.ID), zap.Any("message", fields))
		return nil, nil
	}

	entity := p.entityStore.GetByKey(nodeID, uint32(convertor.ToInteger(key)))
	if entity == nil {
		zap.L().Info("entity not found", zap.String("gatewayId", p.GatewayConfig.ID), zap.Any("message", fields))
		return nil, nil
	}

	messages := make([]*msgML.Message, 0)

	msg := getMessage(p.GatewayConfig.ID, nodeID, entity.SourceID, msgML.TypeSet, timestamp)

	// remove key
	delete(fields, FieldKey)

	for field, value := range fields {
		metricMap := getMetricData(entity.Type, field)
		data := msgML.NewPayload()
		data.Key = field
		data.Value = adjustValueToMyController(entity.Type, field, value)
		data.MetricType = metricMap.MetricType
		if entity.Unit != "" {
			data.Unit = entity.Unit
		} else {
			data.Unit = metricMap.Unit
		}
		data.Labels.Set(LabelKey, convertor.ToString(key))
		data.Labels.Set(LabelType, entity.Type)
		data.Labels.Set(model.LabelReadOnly, metricMap.ReadOnly)
		msg.Payloads = append(msg.Payloads, data)
	}

	// if no state received ignore the data
	if len(msg.Payloads) == 0 {
		return nil, nil
	}
	messages = append(messages, msg)
	return messages, nil
}

// getActionMessage returns the action message to mycontroller
func (p *Provider) getActionMessage(actionType, nodeID string, timestamp time.Time, fields map[string]interface{}) ([]*msgML.Message, error) {
	messages := make([]*msgML.Message, 0)
	msg := getMessage(p.GatewayConfig.ID, nodeID, "", msgML.TypeAction, timestamp)
	data := msgML.NewPayload()
	data.Key = actionType
	msg.Payloads = append(msg.Payloads, data)
	messages = append(messages, msg)
	return messages, nil
}

// getMessage creates a message with common values
func getMessage(gatewayID, nodeID, sourceID, msgType string, timestamp time.Time) *msgML.Message {
	return &msgML.Message{
		GatewayID:  gatewayID,
		NodeID:     nodeID,
		SourceID:   sourceID,
		Timestamp:  timestamp,
		Type:       msgType,
		IsReceived: true,
	}
}

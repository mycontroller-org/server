package tasmota

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/server/v2/pkg/utils/normalize"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	gwPtl "github.com/mycontroller-org/server/v2/plugin/gateway/protocol"
	"go.uber.org/zap"
)

// ToRawMessage converts the message into raw message
func (p *Provider) ToRawMessage(msg *msgTY.Message) (*msgTY.RawMessage, error) {
	if len(msg.Payloads) == 0 {
		return nil, errors.New("there is no payload details on the message")
	}

	// converts exactly the first payload. other payloads are ignored.
	payload := msg.Payloads[0]

	tmMsg := &message{
		Topic:   topicCmnd,
		NodeID:  msg.NodeID,
		Command: payload.Key,
	}

	rawMsg := msgTY.NewRawMessage(false, nil)

	// get command
	switch msg.Type {

	case msgTY.TypeSet: // set payload
		tmMsg.Payload = payload.Value

	case msgTY.TypeRequest: // set empty payload for request type
		tmMsg.Payload = emptyPayload

	case msgTY.TypeAction:
		err := handleActions(p.GatewayConfig, payload.Value, msg, tmMsg)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("this command not implemented: %s", msg.Type)
	}

	// update payload and mqtt topic
	rawMsg.Data = []byte(tmMsg.Payload)
	rawMsg.Others.Set(gwPtl.KeyMqttTopic, []string{tmMsg.toString()}, nil)

	return rawMsg, nil
}

// ProcessReceived converts raw message into message
func (p *Provider) ProcessReceived(rawMsg *msgTY.RawMessage) ([]*msgTY.Message, error) {
	if rawMsg == nil {
		return nil, nil
	}
	// one raw message can contain multiple messages
	messages := make([]*msgTY.Message, 0)

	// topic/node-id/command
	// jktasmota/stat/tasmota_49C88D/STATUS11
	topic, ok := rawMsg.Others.Get(gwPtl.KeyMqttTopic).(string)
	if !ok {
		return nil, fmt.Errorf("unable to get mqtt topic:%v", rawMsg.Others.Get(gwPtl.KeyMqttTopic))
	}
	tSlice := strings.Split(topic, "/")
	if len(tSlice) < 3 {
		zap.L().Error("Invalid message format", zap.Any("rawMessage", rawMsg))
		return nil, nil
	}
	topicSlice := tSlice[len(tSlice)-3:]

	tmMsg := message{
		Topic:   topicSlice[0],
		NodeID:  topicSlice[1],
		Command: topicSlice[2],
	}

	// helper functions

	addIntoMessages := func(msg *msgTY.Message) {
		if msg != nil && len(msg.Payloads) > 0 {
			messages = append(messages, msg)
		}
	}

	addSourcePresentationMessage := func(sourceID string) {
		pl := p.createSourcePresentationPL(sourceID)
		msg := p.createMessage(tmMsg.NodeID, sourceID, msgTY.TypePresentation, *pl)
		addIntoMessages(msg)
	}

	// update metric type and unit
	updateMetricTypeAndUnit := func(key string, pl *msgTY.Payload) {
		mu, found := metricTypeAndUnit[key]
		if found {
			pl.MetricType = mu.Type
			pl.Unit = mu.Unit
		} else {
			pl.MetricType = metricTY.MetricTypeNone
			pl.Unit = metricTY.UnitNone
		}
		if pl.MetricType == metricTY.MetricTypeBinary {
			v := strings.ToLower(pl.Value)
			if v == "on" || v == "1" || v == "true" {
				pl.Value = "1"
			} else {
				pl.Value = "0"
			}
		}
	}

	switch tmMsg.Topic {

	case topicTele:

		// ignore LWT commands
		if tmMsg.Command == cmdLWT {
			return nil, nil
		}

		data := make(map[string]interface{})
		rawMsgBytes, ok := rawMsg.Data.([]byte)
		if !ok {
			zap.L().Error("error on converting to bytes", zap.Any("rawMessage", rawMsg))
			return nil, fmt.Errorf("error on converting to bytes. received: %T", rawMsg.Data)
		}
		err := utils.ToStruct(rawMsgBytes, &data)
		if err != nil {
			return nil, err
		}
		switch tmMsg.Command {
		case cmdResult: // control source messages
			msg := p.createMessage(tmMsg.NodeID, sourceIDControl, msgTY.TypeSet)
			addSourcePresentationMessage(sourceIDControl)

			for key, v := range data {
				// create new payload data
				pl := msgTY.NewPayload()
				pl.Key = key
				pl.Value = converterUtils.ToString(v)
				updateMetricTypeAndUnit(key, &pl)
				msg.Payloads = append(msg.Payloads, pl)
			}
			addIntoMessages(msg)

		case cmdState:
			senControl := p.createMessage(tmMsg.NodeID, sourceIDControl, msgTY.TypeSet)
			addSourcePresentationMessage(sourceIDControl)

			senMemory := p.createMessage(tmMsg.NodeID, sourceIDMemory, msgTY.TypeSet)
			addSourcePresentationMessage(sourceIDMemory)

			for key, v := range data {
				_, ignore := utils.FindItem(teleStateFieldsIgnore, strings.ToLower(key))
				if ignore {
					continue
				}
				if key == keyWifi {
					wiFiData, ok := v.(map[string]interface{})
					if ok {
						senWiFi := p.createMessage(tmMsg.NodeID, sourceIDWiFi, msgTY.TypeSet)
						addSourcePresentationMessage(sourceIDWiFi)
						for wKey, wValue := range wiFiData {
							_, ignore := utils.FindItem(wiFiFieldsIgnore, strings.ToLower(wKey))
							if ignore {
								continue
							}
							pl := msgTY.NewPayload()
							pl.Key = wKey
							pl.Value = converterUtils.ToString(wValue)
							updateMetricTypeAndUnit(wKey, &pl)
							senWiFi.Payloads = append(senWiFi.Payloads, pl)
						}
						addIntoMessages(senWiFi)
					}
				} else if key == keyHSBColor {
					plValue := converterUtils.ToString(v)
					pls := p.getHsbColor(plValue)
					senControl.Payloads = append(senControl.Payloads, pls...)
				} else {
					pl := msgTY.NewPayload()
					pl.Key = key
					pl.Value = converterUtils.ToString(v)
					updateMetricTypeAndUnit(key, &pl)
					if key == keyHeap {
						senMemory.Payloads = append(senMemory.Payloads, pl)
					} else {
						senControl.Payloads = append(senControl.Payloads, pl)
					}
				}

			}
			addIntoMessages(senControl)
			addIntoMessages(senMemory)

		case cmdSensor:
			for k, v := range data {
				dataMap, ok := v.(map[string]interface{})
				if ok {
					msg := p.createMessage(tmMsg.NodeID, k, msgTY.TypeSet)
					addSourcePresentationMessage(k)

					for sK, sV := range dataMap {
						// create new payload data
						pl := msgTY.NewPayload()
						pl.Key = sK
						pl.Value = converterUtils.ToString(sV)
						updateMetricTypeAndUnit(sK, &pl)
						msg.Payloads = append(msg.Payloads, pl)
					}
					addIntoMessages(msg)
				}
			}

		case cmdInfo1, cmdInfo2, cmdInfo3: // node message
			msg := p.getNodeMessage(tmMsg.NodeID, data)
			addIntoMessages(msg)

		default:
			// no action

		}

	case topicStat:
		switch tmMsg.Command {
		case cmdResult:
			data := make(map[string]interface{})
			rawMsgBytes, ok := rawMsg.Data.([]byte)
			if !ok {
				zap.L().Error("error on converting to bytes", zap.Any("rawMessage", rawMsg))
				return nil, fmt.Errorf("error on converting to bytes. received: %T", rawMsg.Data)
			}
			err := utils.ToStruct(rawMsgBytes, &data)
			if err != nil {
				return nil, err
			}
			msg := p.createMessage(tmMsg.NodeID, sourceIDControl, msgTY.TypeSet)
			addSourcePresentationMessage(sourceIDControl)
			for key, v := range data {
				plValue := converterUtils.ToString(v)
				if key == keyHSBColor {
					pls := p.getHsbColor(plValue)
					msg.Payloads = append(msg.Payloads, pls...)
				} else {
					// create new payload data
					pl := msgTY.NewPayload()
					pl.Key = key
					pl.Value = plValue
					updateMetricTypeAndUnit(key, &pl)
					msg.Payloads = append(msg.Payloads, pl)
				}
			}
			addIntoMessages(msg)

		default:
			_, found := utils.FindItem(statusSupported, tmMsg.Command)
			if found {
				// get mesage type
				rawData := make(map[string]map[string]interface{})
				rawMsgBytes, ok := rawMsg.Data.([]byte)
				if !ok {
					zap.L().Error("error on converting to bytes", zap.Any("rawMessage", rawMsg))
					return nil, fmt.Errorf("error on converting to bytes. received: %T", rawMsg.Data)
				}
				err := utils.ToStruct(rawMsgBytes, &rawData)
				if err != nil {
					return nil, err
				}

				header := ""
				var data map[string]interface{}
				for k, v := range rawData {
					header = k
					data = v
				}

				switch header {

				case headerStatus, headerDeviceParameters, headerFirmware, headerNetwork:
					msg := p.getNodeMessage(tmMsg.NodeID, data)
					addIntoMessages(msg)

				case headerLogging: // add all the fields
					msg := p.createMessage(tmMsg.NodeID, sourceIDLogging, msgTY.TypeSet)
					for k, v := range data {
						_, ignore := utils.FindItem(loggingFieldsIgnore, strings.ToLower(k))
						if ignore {
							continue
						}
						pl := msgTY.Payload{
							Key:        k,
							Value:      converterUtils.ToString(v),
							MetricType: metricTY.MetricTypeNone,
						}
						msg.Payloads = append(msg.Payloads, pl)
					}
					addIntoMessages(msg)

					// presentation message
					addSourcePresentationMessage(sourceIDLogging)

				case headerMemory: // update only heap
					msg := p.createMessage(tmMsg.NodeID, sourceIDMemory, msgTY.TypeSet)
					heap, found := data[keyHeap]
					if found {
						pl := msgTY.Payload{
							Key:        keyHeap,
							Value:      converterUtils.ToString(heap),
							MetricType: metricTY.MetricTypeGauge,
						}
						msg.Payloads = append(msg.Payloads, pl)
					}
					addIntoMessages(msg)
					// presentation message
					addSourcePresentationMessage(sourceIDMemory)

				case headerTime:
					msg := p.createMessage(tmMsg.NodeID, sourceIDTime, msgTY.TypeSet)
					addSourcePresentationMessage(sourceIDTime)

					for k, v := range data {
						pl := msgTY.NewPayload()
						pl.Key = k
						pl.Value = converterUtils.ToString(v)
						updateMetricTypeAndUnit(k, &pl)
						msg.Payloads = append(msg.Payloads, pl)
					}
					addIntoMessages(msg)

				case headerSensor:
					getMapFn := func(data interface{}) map[string]interface{} {
						d, ok := data.(map[string]interface{})
						if ok {
							return d
						}
						return nil
					}
					// Update temperature unit
					temperatureUnit := metricTY.UnitCelsius
					if tu, ok := data[keyTemperatureUnit]; ok {
						if tu == "F" {
							temperatureUnit = metricTY.UnitFahrenheit
						}
					}
					for k, v := range data {
						//	value := converterUtils.ToString( v)
						switch k {
						case keyAnalog:
							d := getMapFn(v)
							pls := make([]msgTY.Payload, 0)
							for fName, fValue := range d {
								pl := msgTY.Payload{
									Key:        fName,
									Value:      converterUtils.ToString(fValue),
									MetricType: metricTY.MetricTypeNone,
									Unit:       metricTY.UnitNone,
								}
								pl.Labels = pl.Labels.Init()
								pls = append(pls, pl)
							}

							// field message
							fieldMsg := p.createMessage(tmMsg.NodeID, sourceIDAnalog, msgTY.TypeSet)
							fieldMsg.Payloads = pls
							addIntoMessages(fieldMsg)

							// presentation message
							addSourcePresentationMessage(sourceIDAnalog)

						case keyCounter:
							d := getMapFn(v)
							pls := make([]msgTY.Payload, 0)
							for fName, fValue := range d {
								pl := msgTY.Payload{
									Key:        fName,
									Value:      converterUtils.ToString(fValue),
									MetricType: metricTY.MetricTypeCounter,
									Unit:       metricTY.UnitNone,
								}
								pl.Labels = pl.Labels.Init()
								pls = append(pls, pl)
							}
							// field message
							fieldMsg := p.createMessage(tmMsg.NodeID, sourceIDCounter, msgTY.TypeSet)
							fieldMsg.Payloads = pls
							addIntoMessages(fieldMsg)

							// presentation message
							addSourcePresentationMessage(sourceIDCounter)

						case keyTemperatureUnit:
							// ignore this

						default:
							d := getMapFn(v)
							pls := make([]msgTY.Payload, 0)
							for fName, fValue := range d {
								if fValue == nil {
									continue
								}
								mt, ok := metricTypeAndUnit[fName]
								if !ok {
									mt = payloadMetricTypeUnit{Type: metricTY.MetricTypeNone, Unit: metricTY.UnitNone}
								}

								// update temperature unit
								if fName == keyTemperature {
									mt.Unit = temperatureUnit
								}

								pl := msgTY.NewPayload()
								pl.Key = fName
								pl.Value = converterUtils.ToString(fValue)
								pl.MetricType = mt.Type
								pl.Unit = mt.Unit
								pls = append(pls, pl)
							}
							// field message
							fieldMsg := p.createMessage(tmMsg.NodeID, k, msgTY.TypeSet, pls...)
							addIntoMessages(fieldMsg)

							// presentation message
							addSourcePresentationMessage(k)
						}

					}

				default:
					// print and exit
					zap.L().Debug("no action don for this message", zap.String("header", header), zap.Any("data", data))
				}
			}

		}

	case topicCmnd:
		// ignore

	default:
		// no action
	}

	defer func() {
		zap.L().Debug("update status", zap.Any("messages", messages))
	}()

	return messages, nil
}

// helper functions

// get HsbColor to HsbColor1, HsbColor2, HsbColor3
// input: "HSBColor":"249,0,0"(HsbColor1,2,3)
func (p *Provider) getHsbColor(value string) []msgTY.Payload {
	pls := make([]msgTY.Payload, 0)
	pls = append(pls, msgTY.Payload{Key: keyHSBColor, Value: value, MetricType: metricTY.MetricTypeNone, Unit: metricTY.UnitNone})
	if value != "" && strings.Contains(value, ",") {
		values := strings.Split(value, ",")
		if len(values) == 3 {
			pls = append(pls, msgTY.Payload{Key: keyHSBColor1, Value: values[0], MetricType: metricTY.MetricTypeNone, Unit: metricTY.UnitNone})
			pls = append(pls, msgTY.Payload{Key: keyHSBColor2, Value: values[1], MetricType: metricTY.MetricTypeNone, Unit: metricTY.UnitNone})
			pls = append(pls, msgTY.Payload{Key: keyHSBColor3, Value: values[2], MetricType: metricTY.MetricTypeNone, Unit: metricTY.UnitNone})
		}
	}
	return pls
}

// get node details as payload
func (p *Provider) getNodeMessage(nodeID string, data map[string]interface{}) *msgTY.Message {
	payloads := make([]msgTY.Payload, 0)
	for key, v := range data {
		value := converterUtils.ToString(v)

		// create new payload data
		pl := msgTY.NewPayload()
		pl.Key = normalize.ToSnakeCase(key)
		pl.Value = value
		include := true

		switch key {
		case keyFriendlyName:
			pl.Key = types.FieldName
			names, ok := v.([]interface{})
			if ok {
				if len(names) > 0 {
					pl.Value = converterUtils.ToString(names[0])
				}
			}

		case keyVersion:
			pl.Labels.Set(types.LabelNodeVersion, value)

		case keyCore:
			pl.Labels.Set(types.LabelNodeLibraryVersion, value)

		case keyIPAddress:
			pl.Key = types.FieldIPAddress
			// add host url
			urlPL := msgTY.NewPayload()
			urlPL.Key = types.FieldNodeWebURL
			urlPL.Value = fmt.Sprintf("http://%s", value)
			payloads = append(payloads, urlPL)

		case keyOtaURL, keySDK, keyBuildDateTime, keyCPUFrequency,
			keyHostname, keyMAC, keyRestartReason, keyModule,
			keyFallbackTopic, keyGroupTopic, keyBoot, keyHardware:
			// will be included

		default:
			include = false
		}

		if include {
			payloads = append(payloads, pl)
		}
	}

	if len(payloads) > 0 {
		msg := p.createMessage(nodeID, sourceIDNone, msgTY.TypePresentation)
		msg.Payloads = payloads
		return msg
	} else {
		zap.L().Info("message processed without payload", zap.String("nodeId", nodeID), zap.Any("data", data))
	}
	return nil
}

func (p *Provider) createMessage(nodeID, sourceID, msgType string, pls ...msgTY.Payload) *msgTY.Message {
	msg := msgTY.NewMessage(true)
	msg.GatewayID = p.GatewayConfig.ID
	msg.NodeID = nodeID
	msg.IsAck = false
	msg.Timestamp = time.Now()
	msg.SourceID = sourceID
	msg.Type = msgType
	if len(pls) > 0 {
		msg.Payloads = append(msg.Payloads, pls...)
	}
	return &msg
}

func (p *Provider) createSourcePresentationPL(value string) *msgTY.Payload {
	pl := msgTY.NewPayload()
	pl.Key = types.FieldName
	pl.Value = value
	pl.MetricType = metricTY.MetricTypeNone
	pl.Unit = metricTY.UnitNone
	return &pl
}

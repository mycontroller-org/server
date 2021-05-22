package tasmota

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model"
	msgML "github.com/mycontroller-org/backend/v2/pkg/model/message"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	converterUtils "github.com/mycontroller-org/backend/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/backend/v2/pkg/utils/normalize"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gateway/protocol"
	mtsML "github.com/mycontroller-org/backend/v2/plugin/metrics"
	"go.uber.org/zap"
)

// ToRawMessage converts the message into raw message
func (p *Provider) ToRawMessage(msg *msgML.Message) (*msgML.RawMessage, error) {
	if len(msg.Payloads) == 0 {
		return nil, errors.New("there is no payload details on the message")
	}

	// converts exactly the first payload. other payloads are ignored.
	payload := msg.Payloads[0]

	tmMsg := &message{
		Topic:   topicCmnd,
		NodeID:  msg.NodeID,
		Command: payload.Name,
	}

	rawMsg := msgML.NewRawMessage(false, nil)

	// get command
	switch msg.Type {

	case msgML.TypeSet: // set payload
		tmMsg.Payload = payload.Value

	case msgML.TypeRequest: // set empty payload for request type
		tmMsg.Payload = emptyPayload

	case msgML.TypeAction:
		err := handleActions(p.GatewayConfig, payload.Name, msg, tmMsg)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("this command not implemented: %s", msg.Type)
	}

	// update payload and mqtt topic
	rawMsg.Data = []byte(tmMsg.Payload)
	rawMsg.Others.Set(gwpl.KeyMqttTopic, []string{tmMsg.toString()}, nil)

	return rawMsg, nil
}

// Process converts raw message into message
func (p *Provider) Process(rawMsg *msgML.RawMessage) ([]*msgML.Message, error) {
	// one raw message can contain multiple messages
	messages := make([]*msgML.Message, 0)

	// topic/node-id/command
	// jktasmota/stat/tasmota_49C88D/STATUS11
	topic, ok := rawMsg.Others.Get(gwpl.KeyMqttTopic).(string)
	if !ok {
		return nil, fmt.Errorf("unable to get mqtt topic:%v", rawMsg.Others.Get(gwpl.KeyMqttTopic))
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

	addIntoMessages := func(msg *msgML.Message) {
		if len(msg.Payloads) > 0 {
			messages = append(messages, msg)
		}
	}

	addSourcePresentationMessage := func(sourceID string) {
		pl := p.createSourcePresentationPL(sourceID)
		msg := p.createMessage(tmMsg.NodeID, sourceID, msgML.TypePresentation, *pl)
		addIntoMessages(msg)
	}

	// update metric type and unit
	updateMetricTypeAndUnit := func(key string, pl *msgML.Data) {
		mu, found := metricTypeAndUnit[key]
		if found {
			pl.MetricType = mu.Type
			pl.Unit = mu.Unit
		} else {
			pl.MetricType = mtsML.MetricTypeNone
			pl.Unit = mtsML.UnitNone
		}
		if pl.MetricType == mtsML.MetricTypeBinary {
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
			msg := p.createMessage(tmMsg.NodeID, sourceIDControl, msgML.TypeSet)
			addSourcePresentationMessage(sourceIDControl)

			for key, v := range data {
				// create new payload data
				pl := msgML.NewData()
				pl.Name = key
				pl.Value = converterUtils.ToString(v)
				updateMetricTypeAndUnit(key, &pl)
				msg.Payloads = append(msg.Payloads, pl)
			}
			addIntoMessages(msg)

		case cmdState:
			senControl := p.createMessage(tmMsg.NodeID, sourceIDControl, msgML.TypeSet)
			addSourcePresentationMessage(sourceIDControl)

			senMemory := p.createMessage(tmMsg.NodeID, sourceIDMemory, msgML.TypeSet)
			addSourcePresentationMessage(sourceIDMemory)

			for key, v := range data {
				_, ignore := utils.FindItem(teleStateFieldsIgnore, strings.ToLower(key))
				if ignore {
					continue
				}
				if key == keyWifi {
					wiFiData, ok := v.(map[string]interface{})
					if ok {
						senWiFi := p.createMessage(tmMsg.NodeID, sourceIDWiFi, msgML.TypeSet)
						addSourcePresentationMessage(sourceIDWiFi)
						for wKey, wValue := range wiFiData {
							_, ignore := utils.FindItem(wiFiFieldsIgnore, strings.ToLower(wKey))
							if ignore {
								continue
							}
							pl := msgML.NewData()
							pl.Name = wKey
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
					pl := msgML.NewData()
					pl.Name = key
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
					msg := p.createMessage(tmMsg.NodeID, k, msgML.TypeSet)
					addSourcePresentationMessage(k)

					for sK, sV := range dataMap {
						// create new payload data
						pl := msgML.NewData()
						pl.Name = sK
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
			msg := p.createMessage(tmMsg.NodeID, sourceIDControl, msgML.TypeSet)
			addSourcePresentationMessage(sourceIDControl)
			for key, v := range data {
				plValue := converterUtils.ToString(v)
				if key == keyHSBColor {
					pls := p.getHsbColor(plValue)
					msg.Payloads = append(msg.Payloads, pls...)
				} else {
					// create new payload data
					pl := msgML.NewData()
					pl.Name = key
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
					msg := p.createMessage(tmMsg.NodeID, sourceIDLogging, msgML.TypeSet)
					for k, v := range data {
						_, ignore := utils.FindItem(loggingFieldsIgnore, strings.ToLower(k))
						if ignore {
							continue
						}
						pl := msgML.Data{
							Name:       k,
							Value:      converterUtils.ToString(v),
							MetricType: mtsML.MetricTypeNone,
						}
						msg.Payloads = append(msg.Payloads, pl)
					}
					addIntoMessages(msg)

					// presentation message
					addSourcePresentationMessage(sourceIDLogging)

				case headerMemory: // update only heap
					msg := p.createMessage(tmMsg.NodeID, sourceIDMemory, msgML.TypeSet)
					heap, found := data[keyHeap]
					if found {
						pl := msgML.Data{
							Name:       keyHeap,
							Value:      converterUtils.ToString(heap),
							MetricType: mtsML.MetricTypeGauge,
						}
						msg.Payloads = append(msg.Payloads, pl)
					}
					addIntoMessages(msg)
					// presentation message
					addSourcePresentationMessage(sourceIDMemory)

				case headerTime:
					msg := p.createMessage(tmMsg.NodeID, sourceIDTime, msgML.TypeSet)
					addSourcePresentationMessage(sourceIDTime)

					for k, v := range data {
						pl := msgML.NewData()
						pl.Name = k
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
					temperatureUnit := mtsML.UnitCelsius
					if tu, ok := data[keyTemperatureUnit]; ok {
						if tu == "F" {
							temperatureUnit = mtsML.UnitFahrenheit
						}
					}
					for k, v := range data {
						//	value := converterUtils.ToString( v)
						switch k {
						case keyAnalog:
							d := getMapFn(v)
							pls := make([]msgML.Data, 0)
							for fName, fValue := range d {
								pl := msgML.Data{
									Name:       fName,
									Value:      converterUtils.ToString(fValue),
									MetricType: mtsML.MetricTypeNone,
									Unit:       mtsML.UnitNone,
								}
								pl.Labels = pl.Labels.Init()
								pls = append(pls, pl)
							}

							// field message
							fieldMsg := p.createMessage(tmMsg.NodeID, sourceIDAnalog, msgML.TypeSet)
							fieldMsg.Payloads = pls
							addIntoMessages(fieldMsg)

							// presentation message
							addSourcePresentationMessage(sourceIDAnalog)

						case keyCounter:
							d := getMapFn(v)
							pls := make([]msgML.Data, 0)
							for fName, fValue := range d {
								pl := msgML.Data{
									Name:       fName,
									Value:      converterUtils.ToString(fValue),
									MetricType: mtsML.MetricTypeCounter,
									Unit:       mtsML.UnitNone,
								}
								pl.Labels = pl.Labels.Init()
								pls = append(pls, pl)
							}
							// field message
							fieldMsg := p.createMessage(tmMsg.NodeID, sourceIDCounter, msgML.TypeSet)
							fieldMsg.Payloads = pls
							addIntoMessages(fieldMsg)

							// presentation message
							addSourcePresentationMessage(sourceIDCounter)

						case keyTemperatureUnit:
							// ignore this

						default:
							d := getMapFn(v)
							pls := make([]msgML.Data, 0)
							for fName, fValue := range d {
								if fValue == nil {
									continue
								}
								mt, ok := metricTypeAndUnit[fName]
								if !ok {
									mt = payloadMetricTypeUnit{Type: mtsML.MetricTypeNone, Unit: mtsML.UnitNone}
								}

								// update temperature unit
								if fName == keyTemperature {
									mt.Unit = temperatureUnit
								}

								pl := msgML.NewData()
								pl.Name = fName
								pl.Value = converterUtils.ToString(fValue)
								pl.MetricType = mt.Type
								pl.Unit = mt.Unit
								pls = append(pls, pl)
							}
							// field message
							fieldMsg := p.createMessage(tmMsg.NodeID, k, msgML.TypeSet, pls...)
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
func (p *Provider) getHsbColor(value string) []msgML.Data {
	pls := make([]msgML.Data, 0)
	pls = append(pls, msgML.Data{Name: keyHSBColor, Value: value, MetricType: mtsML.MetricTypeNone, Unit: mtsML.UnitNone})
	if value != "" && strings.Contains(value, ",") {
		values := strings.Split(value, ",")
		if len(values) == 3 {
			pls = append(pls, msgML.Data{Name: keyHSBColor1, Value: values[0], MetricType: mtsML.MetricTypeNone, Unit: mtsML.UnitNone})
			pls = append(pls, msgML.Data{Name: keyHSBColor2, Value: values[1], MetricType: mtsML.MetricTypeNone, Unit: mtsML.UnitNone})
			pls = append(pls, msgML.Data{Name: keyHSBColor3, Value: values[2], MetricType: mtsML.MetricTypeNone, Unit: mtsML.UnitNone})
		}
	}
	return pls
}

// get node details as payload
func (p *Provider) getNodeMessage(nodeID string, data map[string]interface{}) *msgML.Message {
	payloads := make([]msgML.Data, 0)
	for key, v := range data {
		value := converterUtils.ToString(v)

		// create new payload data
		pl := msgML.NewData()
		pl.Name = normalize.ToSnakeCase(key)
		pl.Value = value
		include := true

		switch key {
		case keyFriendlyName:
			pl.Name = model.FieldName
			names, ok := v.([]interface{})
			if ok {
				if len(names) > 0 {
					pl.Value = converterUtils.ToString(names[0])
				}
			}

		case keyVersion:
			pl.Labels.Set(model.LabelNodeVersion, value)

		case keyCore:
			pl.Labels.Set(model.LabelNodeLibraryVersion, value)

		case keyIPAddress:
			pl.Name = model.FieldIPAddress
			// add host url
			urlPL := msgML.NewData()
			urlPL.Name = model.FieldNodeWebURL
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
		msg := p.createMessage(nodeID, sourceIDNone, msgML.TypePresentation)
		msg.Payloads = payloads
		return msg
	}
	return nil
}

func (p *Provider) createMessage(nodeID, sourceID, msgType string, pls ...msgML.Data) *msgML.Message {
	msg := msgML.NewMessage(true)
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

func (p *Provider) createSourcePresentationPL(value string) *msgML.Data {
	pl := msgML.NewData()
	pl.Name = model.FieldName
	pl.Value = value
	pl.MetricType = mtsML.MetricTypeNone
	pl.Unit = mtsML.UnitNone
	return &pl
}

package tasmota

import (
	"errors"
	"fmt"
	"strings"
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gw_protocol"
	mtsml "github.com/mycontroller-org/backend/v2/plugin/metrics"
	"go.uber.org/zap"
)

// ToRawMessage func implementation
func (p *Provider) ToRawMessage(msg *msgml.Message) (*msgml.RawMessage, error) {
	if len(msg.Payloads) == 0 {
		return nil, errors.New("There is no payload details on the message")
	}

	payload := msg.Payloads[0]

	tmMsg := &message{
		Topic:   topicCmnd,
		NodeID:  msg.NodeID,
		Command: payload.Name,
	}

	rawMsg := msgml.NewRawMessage(false, nil)

	// get command
	switch msg.Type {

	case msgml.TypeSet:
		tmMsg.Payload = payload.Value

	case msgml.TypeRequest:
		tmMsg.Payload = emptyPayload

	case msgml.TypeAction:
		err := handleActions(p.GatewayConfig, payload.Name, msg, tmMsg)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("This command not implemented: %s", msg.Type)
	}

	// update payload and mqtt topic
	rawMsg.Data = []byte(tmMsg.Payload)
	rawMsg.Others.Set(gwpl.KeyMqttTopic, []string{tmMsg.toString()}, nil)

	return rawMsg, nil
}

// ToMessage implementation
func (p *Provider) ToMessage(rawMsg *msgml.RawMessage) ([]*msgml.Message, error) {
	messages := make([]*msgml.Message, 0)

	d := make([]string, 0)
	// topic/node-id/command
	// jktasmota/stat/tasmota_49C88D/STATUS11
	topic, ok := rawMsg.Others.Get(gwpl.KeyMqttTopic).(string)
	if !ok {
		return nil, fmt.Errorf("unable to get mqtt topic:%v", rawMsg.Others.Get(gwpl.KeyMqttTopic))
	}
	rData := strings.Split(topic, "/")
	if len(rData) < 3 {
		zap.L().Error("Invalid message format", zap.Any("rawMessage", rawMsg))
		return nil, nil
	}
	d = rData[len(rData)-3:]

	tmMsg := message{
		Topic:   d[0],
		NodeID:  d[1],
		Command: d[2],
	}

	// helper functions
	createMsgFn := func(sensorID, msgType string, pls ...msgml.Data) *msgml.Message {
		msg := msgml.NewMessage(true)
		msg.GatewayID = p.GatewayConfig.ID
		msg.NodeID = tmMsg.NodeID
		msg.IsAck = false
		msg.Timestamp = time.Now()
		msg.SensorID = sensorID
		msg.Type = msgType
		if len(pls) > 0 {
			msg.Payloads = append(msg.Payloads, pls...)
		}
		return &msg
	}

	createSensorPresentationPL := func(value string) *msgml.Data {
		pl := msgml.NewData()
		pl.Name = ml.FieldName
		pl.Value = value
		pl.MetricType = mtsml.MetricTypeNone
		pl.Unit = mtsml.UnitNone
		return &pl
	}

	addIntoMessages := func(msg *msgml.Message) {
		if len(msg.Payloads) > 0 {
			messages = append(messages, msg)
		}
	}

	addSensorPresentationMessage := func(sensorID string) {
		pl := createSensorPresentationPL(sensorID)
		msg := createMsgFn(sensorID, msgml.TypePresentation, *pl)
		addIntoMessages(msg)
	}

	// update metric type and unit
	updateMetricTypeAndUnit := func(key string, pl *msgml.Data) {
		mu, found := metricTypeAndUnit[key]
		if found {
			pl.MetricType = mu.Type
			pl.Unit = mu.Unit
		} else {
			pl.MetricType = mtsml.MetricTypeNone
			pl.Unit = mtsml.UnitNone
		}
		if pl.MetricType == mtsml.MetricTypeBinary {
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
		data := make(map[string]interface{})
		err := ut.ToStruct(rawMsg.Data, &data)
		if err != nil {
			return nil, err
		}
		switch tmMsg.Command {
		case cmdResult: // control sensor messages
			msg := createMsgFn(sensorControl, msgml.TypeSet)
			addSensorPresentationMessage(sensorControl)

			for key, v := range data {
				// create new payload data
				pl := msgml.NewData()
				pl.Name = key
				pl.Value = ut.ToString(v)
				updateMetricTypeAndUnit(key, &pl)
				msg.Payloads = append(msg.Payloads, pl)
			}
			addIntoMessages(msg)

		case cmdState:
			senControl := createMsgFn(sensorControl, msgml.TypeSet)
			addSensorPresentationMessage(sensorControl)

			senMemory := createMsgFn(sensorMemory, msgml.TypeSet)
			addSensorPresentationMessage(sensorMemory)

			for key, v := range data {
				_, ignore := ut.FindItem(teleStateFieldsIgnore, strings.ToLower(key))
				if ignore {
					continue
				}
				if key == keyWifi {
					wiFiData, ok := v.(map[string]interface{})
					if ok {
						senWiFi := createMsgFn(sensorWiFi, msgml.TypeSet)
						addSensorPresentationMessage(sensorWiFi)
						for wKey, wValue := range wiFiData {
							_, ignore := ut.FindItem(wiFiFieldsIgnore, strings.ToLower(wKey))
							if ignore {
								continue
							}
							pl := msgml.NewData()
							pl.Name = wKey
							pl.Value = ut.ToString(wValue)
							updateMetricTypeAndUnit(wKey, &pl)
							senWiFi.Payloads = append(senWiFi.Payloads, pl)
						}
						addIntoMessages(senWiFi)
					}
				} else {
					pl := msgml.NewData()
					pl.Name = key
					pl.Value = ut.ToString(v)
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
					msg := createMsgFn(k, msgml.TypeSet)
					addSensorPresentationMessage(k)

					for sK, sV := range dataMap {
						// create new payload data
						pl := msgml.NewData()
						pl.Name = sK
						pl.Value = ut.ToString(sV)
						updateMetricTypeAndUnit(sK, &pl)
						msg.Payloads = append(msg.Payloads, pl)
					}
					addIntoMessages(msg)
				}
			}

		default:
			// no action

		}

	case topicStat:
		switch tmMsg.Command {
		case cmdResult:
			data := make(map[string]interface{})
			err := ut.ToStruct(rawMsg.Data, &data)
			if err != nil {
				return nil, err
			}
			msg := createMsgFn(sensorControl, msgml.TypeSet)
			addSensorPresentationMessage(sensorControl)
			for key, v := range data {
				// create new payload data
				pl := msgml.NewData()
				pl.Name = key
				pl.Value = ut.ToString(v)
				updateMetricTypeAndUnit(key, &pl)
				msg.Payloads = append(msg.Payloads, pl)
			}
			addIntoMessages(msg)

		default:
			_, found := ut.FindItem(statusSupported, tmMsg.Command)
			if found {
				// get mesage type
				rawData := make(map[string]map[string]interface{})
				err := ut.ToStruct(rawMsg.Data, &rawData)
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
					msg := createMsgFn(sensorIDNone, msgml.TypePresentation)
					for key, v := range data {
						value := ut.ToString(v)

						// create new payload data
						pl := msgml.NewData()
						pl.Name = key
						pl.Value = value
						include := true

						switch key {
						case keyFriendlyName:
							pl.Name = ml.FieldName
							names, ok := v.([]interface{})
							if ok {
								if len(names) > 0 {
									pl.Value = ut.ToString(names[0])
								}
							}

						case keyVersion:
							pl.Labels.Set(ml.LabelNodeVersion, value)

						case keyCore:
							pl.Labels.Set(ml.LabelNodeLibraryVersion, value)

						case keyIPAddress:
							pl.Name = ml.FieldIPAddress
							urlPL := msgml.Data{
								Name:  ml.FieldNodeWebURL,
								Value: fmt.Sprintf("http://%s", value),
							}
							urlPL.Labels = urlPL.Labels.Init()
							msg.Payloads = append(msg.Payloads, urlPL)

						case keyOtaURL, keySDK, keyBuildDateTime, keyCPUFrequency, keyHostname, keyMAC:
							// will be included

						default:
							include = false
						}

						if include {
							msg.Payloads = append(msg.Payloads, pl)
						}
					}
					addIntoMessages(msg)

				case headerLogging: // add all the fields
					msg := createMsgFn(sensorLogging, msgml.TypeSet)
					for k, v := range data {
						_, ignore := ut.FindItem(loggingFieldsIgnore, strings.ToLower(k))
						if ignore {
							continue
						}
						pl := msgml.Data{
							Name:       k,
							Value:      ut.ToString(v),
							MetricType: mtsml.MetricTypeNone,
						}
						msg.Payloads = append(msg.Payloads, pl)
					}
					addIntoMessages(msg)

					// presentation message
					addSensorPresentationMessage(sensorLogging)

				case headerMemory: // update only heap
					msg := createMsgFn(sensorMemory, msgml.TypeSet)
					heap, found := data[keyHeap]
					if found {
						pl := msgml.Data{
							Name:       keyHeap,
							Value:      ut.ToString(heap),
							MetricType: mtsml.MetricTypeGauge,
						}
						msg.Payloads = append(msg.Payloads, pl)
					}
					addIntoMessages(msg)
					// presentation message
					addSensorPresentationMessage(sensorMemory)

				case headerTime:
					msg := createMsgFn(sensorTime, msgml.TypeSet)
					addSensorPresentationMessage(sensorTime)

					for k, v := range data {
						pl := msgml.NewData()
						pl.Name = k
						pl.Value = ut.ToString(v)
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
					temperatureUnit := mtsml.UnitCelsius
					if tu, ok := data[keyTemperatureUnit]; ok {
						if tu == "F" {
							temperatureUnit = mtsml.UnitFahrenheit
						}
					}
					for k, v := range data {
						//	value := ut.ToString( v)
						switch k {
						case keyAnalog:
							d := getMapFn(v)
							pls := make([]msgml.Data, 0)
							for fName, fValue := range d {
								pl := msgml.Data{
									Name:       fName,
									Value:      ut.ToString(fValue),
									MetricType: mtsml.MetricTypeNone,
									Unit:       mtsml.UnitNone,
								}
								pl.Labels = pl.Labels.Init()
								pls = append(pls, pl)
							}

							// field message
							fieldMsg := createMsgFn(sensorAnalog, msgml.TypeSet)
							fieldMsg.Payloads = pls
							addIntoMessages(fieldMsg)

							// presentation message
							addSensorPresentationMessage(sensorAnalog)

						case keyCounter:
							d := getMapFn(v)
							pls := make([]msgml.Data, 0)
							for fName, fValue := range d {
								pl := msgml.Data{
									Name:       fName,
									Value:      ut.ToString(fValue),
									MetricType: mtsml.MetricTypeCounter,
									Unit:       mtsml.UnitNone,
								}
								pl.Labels = pl.Labels.Init()
								pls = append(pls, pl)
							}
							// field message
							fieldMsg := createMsgFn(sensorCounter, msgml.TypeSet)
							fieldMsg.Payloads = pls
							addIntoMessages(fieldMsg)

							// presentation message
							addSensorPresentationMessage(sensorCounter)

						case keyTemperatureUnit:
							// ignore this

						default:
							d := getMapFn(v)
							pls := make([]msgml.Data, 0)
							for fName, fValue := range d {
								if fValue == nil {
									continue
								}
								mt, ok := metricTypeAndUnit[fName]
								if !ok {
									mt = payloadMetricTypeUnit{Type: mtsml.MetricTypeNone, Unit: mtsml.UnitNone}
								}

								// update temperature unit
								if fName == keyTemperature {
									mt.Unit = temperatureUnit
								}

								pl := msgml.NewData()
								pl.Name = fName
								pl.Value = ut.ToString(fValue)
								pl.MetricType = mt.Type
								pl.Unit = mt.Unit
								pls = append(pls, pl)
							}
							// field message
							fieldMsg := createMsgFn(k, msgml.TypeSet, pls...)
							addIntoMessages(fieldMsg)

							// presentation message
							addSensorPresentationMessage(k)
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

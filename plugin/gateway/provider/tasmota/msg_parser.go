package tasmota

import (
	"errors"
	"fmt"
	"strings"
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gateway/protocol"
	mtsml "github.com/mycontroller-org/backend/v2/plugin/metrics"
	"go.uber.org/zap"
)

// ToRawMessage converts the message into raw message
func (p *Provider) ToRawMessage(msg *msgml.Message) (*msgml.RawMessage, error) {
	if len(msg.Payloads) == 0 {
		return nil, errors.New("There is no payload details on the message")
	}

	// converts exactly the first payload. other payloads are ignored.
	payload := msg.Payloads[0]

	tmMsg := &message{
		Topic:   topicCmnd,
		NodeID:  msg.NodeID,
		Command: payload.Name,
	}

	rawMsg := msgml.NewRawMessage(false, nil)

	// get command
	switch msg.Type {

	case msgml.TypeSet: // set payload
		tmMsg.Payload = payload.Value

	case msgml.TypeRequest: // set empty payload for request type
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

// ToMessage converts raw message into message
func (p *Provider) ToMessage(rawMsg *msgml.RawMessage) ([]*msgml.Message, error) {
	// one raw message can contain multiple messages
	messages := make([]*msgml.Message, 0)

	topicSlice := make([]string, 0)
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
	topicSlice = tSlice[len(tSlice)-3:]

	tmMsg := message{
		Topic:   topicSlice[0],
		NodeID:  topicSlice[1],
		Command: topicSlice[2],
	}

	// helper functions

	addIntoMessages := func(msg *msgml.Message) {
		if len(msg.Payloads) > 0 {
			messages = append(messages, msg)
		}
	}

	addSensorPresentationMessage := func(sensorID string) {
		pl := p.createSensorPresentationPL(sensorID)
		msg := p.createMessage(tmMsg.NodeID, sensorID, msgml.TypePresentation, *pl)
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

		// ignore LWT commands
		if tmMsg.Command == cmdLWT {
			return nil, nil
		}

		data := make(map[string]interface{})
		err := ut.ToStruct(rawMsg.Data, &data)
		if err != nil {
			return nil, err
		}
		switch tmMsg.Command {
		case cmdResult: // control sensor messages
			msg := p.createMessage(tmMsg.NodeID, sensorControl, msgml.TypeSet)
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
			senControl := p.createMessage(tmMsg.NodeID, sensorControl, msgml.TypeSet)
			addSensorPresentationMessage(sensorControl)

			senMemory := p.createMessage(tmMsg.NodeID, sensorMemory, msgml.TypeSet)
			addSensorPresentationMessage(sensorMemory)

			for key, v := range data {
				_, ignore := ut.FindItem(teleStateFieldsIgnore, strings.ToLower(key))
				if ignore {
					continue
				}
				if key == keyWifi {
					wiFiData, ok := v.(map[string]interface{})
					if ok {
						senWiFi := p.createMessage(tmMsg.NodeID, sensorWiFi, msgml.TypeSet)
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
					msg := p.createMessage(tmMsg.NodeID, k, msgml.TypeSet)
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
			msg := p.createMessage(tmMsg.NodeID, sensorControl, msgml.TypeSet)
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
					msg := p.createMessage(tmMsg.NodeID, sensorIDNone, msgml.TypePresentation)
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
					msg := p.createMessage(tmMsg.NodeID, sensorLogging, msgml.TypeSet)
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
					msg := p.createMessage(tmMsg.NodeID, sensorMemory, msgml.TypeSet)
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
					msg := p.createMessage(tmMsg.NodeID, sensorTime, msgml.TypeSet)
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
							fieldMsg := p.createMessage(tmMsg.NodeID, sensorAnalog, msgml.TypeSet)
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
							fieldMsg := p.createMessage(tmMsg.NodeID, sensorCounter, msgml.TypeSet)
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
							fieldMsg := p.createMessage(tmMsg.NodeID, k, msgml.TypeSet, pls...)
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

// helper functions
func (p *Provider) createMessage(nodeID, sensorID, msgType string, pls ...msgml.Data) *msgml.Message {
	msg := msgml.NewMessage(true)
	msg.GatewayID = p.GatewayConfig.ID
	msg.NodeID = nodeID
	msg.IsAck = false
	msg.Timestamp = time.Now()
	msg.SensorID = sensorID
	msg.Type = msgType
	if len(pls) > 0 {
		msg.Payloads = append(msg.Payloads, pls...)
	}
	return &msg
}

func (p *Provider) createSensorPresentationPL(value string) *msgml.Data {
	pl := msgml.NewData()
	pl.Name = ml.FieldName
	pl.Value = value
	pl.MetricType = mtsml.MetricTypeNone
	pl.Unit = mtsml.UnitNone
	return &pl
}

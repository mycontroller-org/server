package tasmota

import (
	"fmt"
	"strings"
	"time"

	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	msgml "github.com/mycontroller-org/backend/v2/pkg/model/message"
	mtrml "github.com/mycontroller-org/backend/v2/pkg/model/metric"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
	gwpl "github.com/mycontroller-org/backend/v2/plugin/gw_protocol"
	"go.uber.org/zap"
)

// ToRawMessage func implementation
func (p *Provider) ToRawMessage(msg *msgml.Message) (*msgml.RawMessage, error) {
	return nil, nil
}

// ToMessage implementation
func (p *Provider) ToMessage(rawMsg *msgml.RawMessage) ([]*msgml.Message, error) {
	messages := make([]*msgml.Message, 0)

	d := make([]string, 0)
	// topic/node-id/command
	// jktasmota/stat/tasmota_49C88D/STATUS11
	rData := strings.Split(string(rawMsg.Others.Get(gwpl.KeyMqttTopic).(string)), "/")
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

	isJSONData := false

	if tmMsg.Topic == topicTele || tmMsg.Topic == topicStat {
		isJSONData = true
	}

	createMsgFn := func(sensorID, msgType string, pls ...msgml.Data) *msgml.Message {
		msg := msgml.NewMessage(true)
		msg.GatewayID = p.GWConfig.ID
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
		pl.MetricType = mtrml.MetricTypeNone
		pl.Unit = mtrml.UnitNone
		return &pl
	}

	addIntoMessages := func(msg *msgml.Message) {
		if len(msg.Payloads) > 0 {
			messages = append(messages, msg)
		}
	}

	if isJSONData {
		switch {

		case ut.IsExists(cmdWithHeader, tmMsg.Command):
			// get mesage type
			out := make(map[string]map[string]interface{})
			err := toStruct(rawMsg.Data, &out)
			if err != nil {
				return nil, err
			}

			header := ""
			var data map[string]interface{}
			for k, v := range out {
				header = k
				data = v
			}

			switch header {

			case headerStatus, headerDeviceParameters, headerFirmware, headerNetwork:
				msg := createMsgFn(sensorIDNone, msgml.TypePresentation)
				for key, v := range data {
					value := fmt.Sprintf("%v", v)

					// create new payload data
					pl := msgml.NewData()
					pl.Name = key
					pl.Value = value
					include := true

					switch key {
					case keyFriendlyName:
						pl.Name = ml.FieldName
						names := v.([]interface{})
						if len(names) > 0 {
							pl.Value = fmt.Sprintf("%v", names[0])
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
					pl := msgml.Data{
						Name:       k,
						Value:      fmt.Sprintf("%v", v),
						MetricType: mtrml.MetricTypeNone,
					}
					msg.Payloads = append(msg.Payloads, pl)
				}
				addIntoMessages(msg)

				// presentation message
				prsPL := createSensorPresentationPL(sensorLogging)
				prsMsg := createMsgFn(sensorLogging, msgml.TypePresentation, *prsPL)
				addIntoMessages(prsMsg)

			case headerMemory: // update only heap
				msg := createMsgFn(sensorMemory, msgml.TypeSet)
				heap, found := data[keyHeap]
				if found {
					pl := msgml.Data{
						Name:       keyHeap,
						Value:      fmt.Sprintf("%v", heap),
						MetricType: mtrml.MetricTypeGauge,
					}
					msg.Payloads = append(msg.Payloads, pl)
				}
				addIntoMessages(msg)
				// presentation message
				prsPL := createSensorPresentationPL(sensorMemory)
				prsMsg := createMsgFn(sensorMemory, msgml.TypePresentation, *prsPL)
				addIntoMessages(prsMsg)

			case headerSensor:
				getMapFn := func(data interface{}) map[string]interface{} {
					d, ok := data.(map[string]interface{})
					if ok {
						return d
					}
					return nil
				}
				// Update temperature unit
				temperatureUnit := mtrml.UnitCelsius
				if tu, ok := data[keyTemperatureUnit]; ok {
					if tu == "F" {
						temperatureUnit = mtrml.UnitFahrenheit
					}
				}
				for k, v := range data {
					//	value := fmt.Sprintf("%v", v)
					switch k {
					case keyAnalog:
						d := getMapFn(v)
						pls := make([]msgml.Data, 0)
						for fName, fValue := range d {
							pl := msgml.Data{
								Name:       fName,
								Value:      fmt.Sprintf("%v", fValue),
								MetricType: mtrml.MetricTypeNone,
								Unit:       mtrml.UnitNone,
							}
							pl.Labels = pl.Labels.Init()
							pls = append(pls, pl)
						}

						// field message
						fieldMsg := createMsgFn(sensorAnalog, msgml.TypeSet)
						fieldMsg.Payloads = pls
						addIntoMessages(fieldMsg)

						// presentation message
						prsPL := createSensorPresentationPL(sensorAnalog)
						prsMsg := createMsgFn(sensorAnalog, msgml.TypePresentation, *prsPL)
						addIntoMessages(prsMsg)

					case keyCounter:
						d := getMapFn(v)
						pls := make([]msgml.Data, 0)
						for fName, fValue := range d {
							pl := msgml.Data{
								Name:       fName,
								Value:      fmt.Sprintf("%v", fValue),
								MetricType: mtrml.MetricTypeCounter,
								Unit:       mtrml.UnitNone,
							}
							pl.Labels = pl.Labels.Init()
							pls = append(pls, pl)
						}
						// field message
						fieldMsg := createMsgFn(sensorCounter, msgml.TypeSet)
						fieldMsg.Payloads = pls
						addIntoMessages(fieldMsg)

						// presentation message
						prsPL := createSensorPresentationPL(sensorCounter)
						prsMsg := createMsgFn(sensorCounter, msgml.TypePresentation, *prsPL)
						addIntoMessages(prsMsg)

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
								mt = payloadMetricTypeUnit{Type: mtrml.MetricTypeNone, Unit: mtrml.UnitNone}
							}

							// update temperature unit
							if fName == keyTemperature {
								mt.Unit = temperatureUnit
							}

							pl := msgml.NewData()
							pl.Name = fName
							pl.Value = fmt.Sprintf("%v", fValue)
							pl.MetricType = mt.Type
							pl.Unit = mt.Unit
							pls = append(pls, pl)
						}
						// field message
						fieldMsg := createMsgFn(k, msgml.TypeSet, pls...)
						addIntoMessages(fieldMsg)

						// presentation message
						prsPL := createSensorPresentationPL(k)
						prsMsg := createMsgFn(k, msgml.TypePresentation, *prsPL)
						addIntoMessages(prsMsg)
					}

				}

			default:
				// print and exit
				zap.L().Debug("*** no action don for this message", zap.String("header", header), zap.Any("data", data))
			}
		}

	}

	defer func() {
		zap.L().Debug("update status", zap.Any("messages", messages))
	}()

	return messages, nil
}

// {"Time":"2020-09-10T12:18:55","COUNTER":{"C1":0,"C3":0},"ANALOG":{"A0":4},"DHT11-00":{"Temperature":null,"Humidity":null,"DewPoint":null},"AM2301-01":{"Temperature":null,"Humidity":null,"DewPoint":null},"TempUnit":"C"}
// topic : jktasmota/tele/tasmota_49C88D/SENSOR

// {"Time":"2020-09-10T12:18:55","Uptime":"0T06:15:09","UptimeSec":22509,"Heap":25,"SleepMode":"Dynamic","Sleep":50,"LoadAvg":24,"MqttCount":1,"POWER1":"OFF","POWER2":"OFF","Dimmer":0,"Fade":"OFF","Speed":1,"LedTable":"ON","Wifi":{"AP":1,"SSId":"jee","BSSId":"C4:E9:84:5A:DD:CC","Channel":5,"RSSI":74,"Signal":-63,"LinkCount":1,"Downtime":"0T00:00:03"}}
// topic : jktasmota/tele/tasmota_49C88D/STATE

// {"StatusSTS":{"Time":"2020-09-10T12:18:02","Uptime":"0T06:14:16","UptimeSec":22456,"Heap":21,"SleepMode":"Dynamic","Sleep":50,"LoadAvg":23,"MqttCount":1,"POWER1":"OFF","POWER2":"OFF","Dimmer":0,"Fade":"OFF","Speed":1,"LedTable":"ON","Wifi":{"AP":1,"SSId":"jee","BSSId":"C4:E9:84:5A:DD:CC","Channel":5,"RSSI":80,"Signal":-60,"LinkCount":1,"Downtime":"0T00:00:03"}}}
// topic : jktasmota/stat/tasmota_49C88D/STATUS11

// {"StatusSNS":{"Time":"2020-09-10T12:18:02","COUNTER":{"C1":0,"C3":0},"ANALOG":{"A0":4},"DHT11-00":{"Temperature":null,"Humidity":null,"DewPoint":null},"AM2301-01":{"Temperature":null,"Humidity":null,"DewPoint":null},"TempUnit":"C"}}
// topic : jktasmota/stat/tasmota_49C88D/STATUS10

// {"StatusTIM":{"UTC":"2020-09-10T11:18:02","Local":"2020-09-10T12:18:02","StartDST":"2020-03-29T02:00:00","EndDST":"2020-10-25T03:00:00","Timezone":"+01:00","Sunrise":"06:20","Sunset":"19:13"}}
// topic : jktasmota/stat/tasmota_49C88D/STATUS7

// {"StatusMQT":{"MqttHost":"enveedu.mycontroller.org","MqttPort":2883,"MqttClientMask":"DVES_%06X","MqttClient":"DVES_49C88D","MqttUser":"DVES_USER","MqttCount":1,"MAX_PACKET_SIZE":1200,"KEEPALIVE":30}}
// topic : jktasmota/stat/tasmota_49C88D/STATUS6

// {"StatusNET":{"Hostname":"tasmota_49C88D-2189","IPAddress":"192.168.21.113","Gateway":"192.168.21.1","Subnetmask":"255.255.255.0","DNSServer":"192.168.21.1","Mac":"60:01:94:49:C8:8D","Webserver":2,"WifiConfig":4,"WifiPower":17.0}}
// topic : jktasmota/stat/tasmota_49C88D/STATUS5

// {"StatusMEM":{"ProgramSize":595,"Free":408,"Heap":22,"ProgramFlashSize":4096,"FlashSize":4096,"FlashChipId":"1640EF","FlashFrequency":40,"FlashMode":3,"Features":["00000809","8FDAE797","04368001","000000CD","010013C0","C000F981","00004004","00000000"],"Drivers":"1,2,3,4,5,6,7,8,9,10,12,16,18,19,20,21,22,24,26,27,29,30,35,37","Sensors":"1,2,3,4,5,6"}}
// topic : jktasmota/stat/tasmota_49C88D/STATUS4

// {"StatusLOG":{"SerialLog":0,"WebLog":2,"MqttLog":0,"SysLog":0,"LogHost":"","LogPort":514,"SSId":["jee",""],"TelePeriod":300,"Resolution":"558180C0","SetOption":["00008009","2805C8000100060000005A00000000000000","00000000","00006000","00000000"]}}
// topic : jktasmota/stat/tasmota_49C88D/STATUS3

// {"StatusFWR":{"Version":"8.5.0(tasmota)","BuildDateTime":"2020-09-09T11:41:02","Boot":31,"Core":"2_7_4_1","SDK":"2.2.2-dev(38a443e)","CpuFrequency":80,"Hardware":"ESP8266EX","CR":"366/699"}}
// topic : jktasmota/stat/tasmota_49C88D/STATUS2

// {"StatusPRM":{"Baudrate":115200,"SerialConfig":"8N1","GroupTopic":"tasmotas","OtaUrl":"http://ota.tasmota.com/tasmota/release/tasmota.bin","RestartReason":"External System","Uptime":"0T06:14:15","StartupUTC":"2020-09-10T05:03:46","Sleep":50,"CfgHolder":4617,"BootCount":25,"BCResetTime":"2020-09-09T18:01:31","SaveCount":86,"SaveAddress":"F6000"}}
// topic : jktasmota/stat/tasmota_49C88D/STATUS1

// {"Status":{"Module":0,"DeviceName":"Tasmota","FriendlyName":["Tasmota","Tasmota2"],"Topic":"tasmota_49C88D","ButtonTopic":"0","Power":0,"PowerOnState":3,"LedState":1,"LedMask":"FFFF","SaveData":1,"SaveState":1,"SwitchTopic":"0","SwitchMode":[0,0,0,0,0,0,0,0],"ButtonRetain":0,"SwitchRetain":0,"SensorRetain":0,"PowerRetain":0}}
// topic : jktasmota/stat/tasmota_49C88D/STATUS

// 0
// topic : jktasmota/cmnd/tasmota_49C88D/Status

// OFF
// topic : jktasmota/stat/tasmota_49C88D/POWER1

// {"POWER1":"OFF"}
// topic : jktasmota/stat/tasmota_49C88D/RESULT

// 0
// topic : jktasmota/cmnd/tasmota_49C88D/Power1

// ON
// topic : jktasmota/stat/tasmota_49C88D/POWER1

// {"POWER1":"ON"}
// topic : jktasmota/stat/tasmota_49C88D/RESULT

// 1
// topic : jktasmota/cmnd/tasmota_49C88D/Power1

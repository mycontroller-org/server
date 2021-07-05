package philipshue

import (
	"fmt"
	"time"

	"github.com/amimof/huego"
	"github.com/mycontroller-org/server/v2/pkg/model"
	msgML "github.com/mycontroller-org/server/v2/pkg/model/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	metricType "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
	"go.uber.org/zap"
)

func (p *Provider) getUpdate() {
	p.updateLights()
	p.updateSensors()
}

func (p *Provider) updateLights() {
	lights, err := p.bridge.GetLights()
	if err != nil {
		zap.L().Error("error on fetching lights", zap.Error(err))
		return
	}

	for _, light := range lights {
		p.updateLight(&light)
	}
}

func (p *Provider) updateLight(light *huego.Light) {
	nodeID := fmt.Sprintf("light_%v", light.ID)
	// update node presentation message
	presnMsg := p.getPresentationMsg(nodeID, "")

	nodeData := msgML.NewPayload()
	nodeData.Key = model.FieldName
	nodeData.Value = light.Name
	nodeData.Labels.Set(model.LabelNodeVersion, light.SwVersion)
	nodeData.Labels.Set("unique_id", light.UniqueID)
	nodeData.Labels.Set("model_id", light.ModelID)

	nodeData.Others.Set("type", light.Type, nil)
	nodeData.Others.Set("id", light.ID, nil)
	nodeData.Others.Set("manufacturer_name", light.ManufacturerName, nil)
	nodeData.Others.Set("model_id", light.ModelID, nil)
	nodeData.Others.Set("product_id", light.ProductID, nil)
	nodeData.Others.Set("sw_config_id", light.SwConfigID, nil)
	nodeData.Others.Set("unique_id", light.UniqueID, nil)
	nodeData.Others.Set("sw_version", light.SwVersion, nil)

	presnMsg.Payloads = append(presnMsg.Payloads, nodeData)
	err := p.postMsg(presnMsg)
	if err != nil {
		zap.L().Error("error on posting message", zap.Error(err))
		return
	}

	// update source presentation messages
	sourceMsg := p.getPresentationMsg(nodeID, SourceState)
	stateMsgData := msgML.NewPayload()
	stateMsgData.Key = model.FieldName
	stateMsgData.Value = "State"

	sourceMsg.Payloads = append(sourceMsg.Payloads, stateMsgData)
	err = p.postMsg(sourceMsg)
	if err != nil {
		zap.L().Error("error on posting message", zap.Error(err))
		return
	}

	// update state fields
	stateMsg := p.getMsg(nodeID, SourceState)
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldPower, light.State.On, metricType.MetricTypeBinary, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldBrightness, light.State.Bri, metricType.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldHue, light.State.Hue, metricType.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldSaturation, light.State.Sat, metricType.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldColorTemperature, light.State.Ct, metricType.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldAlert, light.State.Alert, metricType.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldEffect, light.State.Effect, metricType.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldColorMode, light.State.ColorMode, metricType.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldReachable, light.State.Reachable, metricType.MetricTypeNone, true))
	err = p.postMsg(stateMsg)
	if err != nil {
		zap.L().Error("error on posting message", zap.Error(err))
		return
	}
}

func (p *Provider) updateSensors() {
	sensors, err := p.bridge.GetSensors()
	if err != nil {
		zap.L().Error("error on fetching sensors", zap.Error(err))
		return
	}

	for _, sensor := range sensors {
		p.updateSensor(&sensor)
	}
}

func (p *Provider) updateSensor(sensor *huego.Sensor) {
	nodeID := fmt.Sprintf("sensor_%v", sensor.ID)
	// update node presentation message
	presnMsg := p.getPresentationMsg(nodeID, "")

	nodeData := msgML.NewPayload()
	nodeData.Key = model.FieldName
	nodeData.Value = sensor.Name
	nodeData.Labels.Set(model.LabelNodeVersion, sensor.SwVersion)
	nodeData.Labels.Set("unique_id", sensor.UniqueID)
	nodeData.Labels.Set("model_id", sensor.ModelID)

	nodeData.Others.Set("type", sensor.Type, nil)
	nodeData.Others.Set("id", sensor.ID, nil)
	nodeData.Others.Set("manufacturer_name", sensor.ManufacturerName, nil)
	nodeData.Others.Set("model_id", sensor.ModelID, nil)
	nodeData.Others.Set("unique_id", sensor.UniqueID, nil)
	nodeData.Others.Set("sw_version", sensor.SwVersion, nil)

	presnMsg.Payloads = append(presnMsg.Payloads, nodeData)
	err := p.postMsg(presnMsg)
	if err != nil {
		zap.L().Error("error on posting message", zap.Error(err))
		return
	}

	// update source presentation messages
	sourceStateMsg := p.getPresentationMsg(nodeID, SourceState)
	stateMsgData := msgML.NewPayload()
	stateMsgData.Key = model.FieldName
	stateMsgData.Value = "State"

	sourceStateMsg.Payloads = append(sourceStateMsg.Payloads, stateMsgData)
	err = p.postMsg(sourceStateMsg)
	if err != nil {
		zap.L().Error("error on posting message", zap.Error(err))
		return
	}

	// update state fields
	stateMsg := p.getMsg(nodeID, SourceState)
	for key, value := range sensor.State {
		stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(key, value, metricType.MetricTypeNone, false))
	}

	err = p.postMsg(stateMsg)
	if err != nil {
		zap.L().Error("error on posting message", zap.Error(err))
		return
	}

	// update config fields
	sourceConfigMsg := p.getPresentationMsg(nodeID, SourceConfig)
	configMsgData := msgML.NewPayload()
	configMsgData.Key = model.FieldName
	configMsgData.Value = "Config"

	sourceConfigMsg.Payloads = append(sourceConfigMsg.Payloads, configMsgData)
	err = p.postMsg(sourceConfigMsg)
	if err != nil {
		zap.L().Error("error on posting message", zap.Error(err))
		return
	}

	// update config fields
	configMsg := p.getMsg(nodeID, SourceConfig)
	for key, value := range sensor.State {
		configMsg.Payloads = append(stateMsg.Payloads, p.getPayload(key, value, metricType.MetricTypeNone, true))
	}

	err = p.postMsg(configMsg)
	if err != nil {
		zap.L().Error("error on posting message", zap.Error(err))
		return
	}
}

func (p *Provider) updateBridgeDetails() {
	brCfg, err := p.bridge.GetConfig()
	if err != nil {
		zap.L().Error("error on getting bridge configuration", zap.Error(err))
		// mark the gateway is in error state
		state := model.State{
			Status:  model.StatusError,
			Message: err.Error(),
			Since:   time.Now(),
		}
		busUtils.SetGatewayState(p.GatewayConfig.ID, state)
		return
	}

	// update node presentation message
	presnMsg := p.getPresentationMsg(NodeBridge, "")

	nodeData := msgML.NewPayload()
	nodeData.Key = model.FieldName
	nodeData.Value = brCfg.Name
	nodeData.Labels.Set(model.LabelNodeVersion, brCfg.SwVersion)
	nodeData.Labels.Set("bridge_id", brCfg.BridgeID)
	nodeData.Labels.Set("model_id", brCfg.ModelID)

	dataMap := utils.StructToMap(brCfg)
	for key, value := range dataMap {
		nodeData.Others.Set(key, value, nil)
	}

	presnMsg.Payloads = append(presnMsg.Payloads, nodeData)
	err = p.postMsg(presnMsg)
	if err != nil {
		zap.L().Error("error on posting message", zap.Error(err))
		return
	}
}

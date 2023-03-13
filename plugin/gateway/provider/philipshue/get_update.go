package philipshue

import (
	"fmt"
	"time"

	"github.com/amimof/huego"
	"github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	"go.uber.org/zap"
)

func (p *Provider) getUpdate() {
	p.updateLights()
	p.updateSensors()
}

func (p *Provider) updateLights() {
	lights, err := p.bridge.GetLights()
	if err != nil {
		p.logger.Error("error on fetching lights", zap.Error(err))
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

	nodeData := msgTY.NewPayload()
	nodeData.Key = types.FieldName
	nodeData.SetValue(light.Name)
	nodeData.Labels.Set(types.LabelNodeVersion, light.SwVersion)
	nodeData.Labels.Set("unique_id", light.UniqueID)
	nodeData.Labels.Set("model_id", light.ModelID)

	nodeData.Others.Set("type", light.Type, nil)
	nodeData.Others.Set("id", light.ID, nil)
	nodeData.Others.Set("manufacturer_name", light.ManufacturerName, nil)
	nodeData.Others.Set("model_id", light.ModelID, nil)
	nodeData.Others.Set("product_name", light.ProductName, nil)
	nodeData.Others.Set("sw_config_id", light.SwConfigID, nil)
	nodeData.Others.Set("unique_id", light.UniqueID, nil)
	nodeData.Others.Set("sw_version", light.SwVersion, nil)

	presnMsg.Payloads = append(presnMsg.Payloads, nodeData)
	err := p.postMsg(presnMsg)
	if err != nil {
		p.logger.Error("error on posting message", zap.Error(err))
		return
	}

	// update source presentation messages
	sourceMsg := p.getPresentationMsg(nodeID, SourceState)
	stateMsgData := msgTY.NewPayload()
	stateMsgData.Key = types.FieldName
	stateMsgData.Value = "State"

	sourceMsg.Payloads = append(sourceMsg.Payloads, stateMsgData)
	err = p.postMsg(sourceMsg)
	if err != nil {
		p.logger.Error("error on posting message", zap.Error(err))
		return
	}

	// update state fields
	stateMsg := p.getMsg(nodeID, SourceState)
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldPower, light.State.On, metricTY.MetricTypeBinary, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldBrightness, light.State.Bri, metricTY.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldHue, light.State.Hue, metricTY.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldSaturation, light.State.Sat, metricTY.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldColorTemperature, light.State.Ct, metricTY.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldAlert, light.State.Alert, metricTY.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldEffect, light.State.Effect, metricTY.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldColorMode, light.State.ColorMode, metricTY.MetricTypeNone, false))
	stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(FieldReachable, light.State.Reachable, metricTY.MetricTypeNone, true))
	err = p.postMsg(stateMsg)
	if err != nil {
		p.logger.Error("error on posting message", zap.Error(err))
		return
	}
}

func (p *Provider) updateSensors() {
	sensors, err := p.bridge.GetSensors()
	if err != nil {
		p.logger.Error("error on fetching sensors", zap.Error(err))
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

	nodeData := msgTY.NewPayload()
	nodeData.Key = types.FieldName
	nodeData.SetValue(sensor.Name)
	nodeData.Labels.Set(types.LabelNodeVersion, sensor.SwVersion)
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
		p.logger.Error("error on posting message", zap.Error(err))
		return
	}

	// update source presentation messages
	sourceStateMsg := p.getPresentationMsg(nodeID, SourceState)
	stateMsgData := msgTY.NewPayload()
	stateMsgData.Key = types.FieldName
	stateMsgData.Value = "State"

	sourceStateMsg.Payloads = append(sourceStateMsg.Payloads, stateMsgData)
	err = p.postMsg(sourceStateMsg)
	if err != nil {
		p.logger.Error("error on posting message", zap.Error(err))
		return
	}

	// update state fields
	stateMsg := p.getMsg(nodeID, SourceState)
	for key, value := range sensor.State {
		stateMsg.Payloads = append(stateMsg.Payloads, p.getPayload(key, value, metricTY.MetricTypeNone, false))
	}

	err = p.postMsg(stateMsg)
	if err != nil {
		p.logger.Error("error on posting message", zap.Error(err))
		return
	}

	// update config fields
	sourceConfigMsg := p.getPresentationMsg(nodeID, SourceConfig)
	configMsgData := msgTY.NewPayload()
	configMsgData.Key = types.FieldName
	configMsgData.Value = "Config"

	sourceConfigMsg.Payloads = append(sourceConfigMsg.Payloads, configMsgData)
	err = p.postMsg(sourceConfigMsg)
	if err != nil {
		p.logger.Error("error on posting message", zap.Error(err))
		return
	}

	// update config fields
	configMsg := p.getMsg(nodeID, SourceConfig)
	for key, value := range sensor.State {
		configMsg.Payloads = append(stateMsg.Payloads, p.getPayload(key, value, metricTY.MetricTypeNone, true))
	}

	err = p.postMsg(configMsg)
	if err != nil {
		p.logger.Error("error on posting message", zap.Error(err))
		return
	}
}

func (p *Provider) updateBridgeDetails() {
	brCfg, err := p.bridge.GetConfig()
	if err != nil {
		p.logger.Error("error on getting bridge configuration", zap.Error(err))
		// mark the gateway is in error state
		state := types.State{
			Status:  types.StatusError,
			Message: err.Error(),
			Since:   time.Now(),
		}
		busUtils.SetGatewayState(p.logger, p.bus, p.GatewayConfig.ID, state)
		return
	}

	// update node presentation message
	presnMsg := p.getPresentationMsg(NodeBridge, "")

	nodeData := msgTY.NewPayload()
	nodeData.Key = types.FieldName
	nodeData.SetValue(brCfg.Name)
	nodeData.Labels.Set(types.LabelNodeVersion, brCfg.SwVersion)
	nodeData.Labels.Set("bridge_id", brCfg.BridgeID)
	nodeData.Labels.Set("model_id", brCfg.ModelID)

	dataMap := utils.StructToMap(brCfg)
	for key, value := range dataMap {
		nodeData.Others.Set(key, value, nil)
	}

	presnMsg.Payloads = append(presnMsg.Payloads, nodeData)
	err = p.postMsg(presnMsg)
	if err != nil {
		p.logger.Error("error on posting message", zap.Error(err))
		return
	}
}

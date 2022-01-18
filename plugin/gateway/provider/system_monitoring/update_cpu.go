package systemmonitoring

import (
	"fmt"

	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	"github.com/mycontroller-org/server/v2/plugin/gateway/provider/system_monitoring/config"
	"github.com/shirou/gopsutil/v3/cpu"
	"go.uber.org/zap"
)

func (p *Provider) updateCPU() {
	if !p.HostConfig.CPU.CPUDisabled {
		// send presentation message
		infoList, err := cpu.Info()
		if err != nil {
			zap.L().Error("error on getting cpu info", zap.Error(err))
			return
		}
		if len(infoList) == 0 {
			return
		}
		sendCPUInfo(p, &infoList[0], -1, len(infoList))

		// update cpu usage
		usage, err := cpu.Percent(0, false)
		if err != nil {
			zap.L().Error("error on getting cpu usage", zap.Error(err))
			return
		}
		if len(usage) == 0 {
			return
		}

		msg := p.getMsg(config.SourceTypeCPU)
		msg.Payloads = append(msg.Payloads, p.getData("used_percent", usage[0], metricTY.MetricTypeGaugeFloat, metricTY.UnitPercent, true))

		err = p.postMsg(&msg)
		if err != nil {
			zap.L().Error("error on posting msg", zap.Error(err))
			return
		}
	}

	if !p.HostConfig.CPU.PerCPUDisabled {
		// send presentation message
		infoList, err := cpu.Info()
		if err != nil {
			zap.L().Error("error on getting cpu info", zap.Error(err))
			return
		}
		if len(infoList) == 0 {
			return
		}
		for _, cpuInfo := range infoList {
			sendCPUInfo(p, &cpuInfo, int(cpuInfo.CPU), len(infoList))
		}

		// update cpu usage
		usageList, err := cpu.Percent(0, true)
		if err != nil {
			zap.L().Error("error on getting cpu usage", zap.Error(err))
			return
		}
		if len(usageList) == 0 {
			return
		}

		for index, usage := range usageList {
			sourceID := fmt.Sprintf("%s_%v", config.SourceTypeCPU, index)
			msg := p.getMsg(sourceID)
			msg.Payloads = append(msg.Payloads, p.getData("used_percent", usage, metricTY.MetricTypeGaugeFloat, metricTY.UnitPercent, true))

			err = p.postMsg(&msg)
			if err != nil {
				zap.L().Error("error on posting msg", zap.Error(err))
				return
			}
		}

	}

}

func sendCPUInfo(p *Provider, cpuInfo *cpu.InfoStat, cpuIndex, cpuCount int) {
	sourceID := config.SourceTypeCPU
	sourceName := "CPU"

	if cpuIndex > -1 {
		sourceID = fmt.Sprintf("%s_%v", config.SourceTypeCPU, cpuIndex)
		sourceName = fmt.Sprintf("CPU_%v", cpuIndex)
	}

	// presentation message
	presentMsg := p.getSourcePresentationMsg(sourceID, sourceName)

	var data msgTY.Payload
	if len(presentMsg.Payloads) > 0 {
		data = presentMsg.Payloads[0]
	} else {
		data = p.getData(sourceID, sourceName, metricTY.MetricTypeNone, metricTY.UnitNone, true)
		presentMsg.Payloads = append(presentMsg.Payloads, data)
	}

	// update cpu details
	// include cpu details
	infoMap := utils.StructToMap(&cpuInfo)
	for key, value := range infoMap {
		data.Others.Set(key, value, nil)
	}
	data.Others.Set("cpu_count", cpuCount, nil)

	// update lables
	data.Labels.Set("vendor_id", cpuInfo.VendorID)
	data.Labels.Set("frequency", convertor.ToString(cpuInfo.Mhz))
	data.Labels.Set("model_name", convertor.ToString(cpuInfo.ModelName))
	data.Labels.Set("cpu_count", convertor.ToString(cpuCount))

	// update info
	presentMsg.Payloads[0] = data

	err := p.postMsg(&presentMsg)
	if err != nil {
		zap.L().Error("error on posting presentation msg", zap.Error(err))
		return
	}
}

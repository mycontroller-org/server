package systemmonitoring

import (
	"github.com/mycontroller-org/backend/v2/plugin/gateway/provider/system_monitoring/config"
	metricsML "github.com/mycontroller-org/backend/v2/plugin/metrics"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
)

func (p *Provider) updateMemory() {
	// presentation message
	presentMsg := p.getSourcePresentationMsg(config.SourceTypeMemory, "Memory")

	err := p.postMsg(&presentMsg)
	if err != nil {
		zap.L().Error("error on posting presentation msg", zap.Error(err))
		return
	}

	// data message
	msg := p.getMsg(config.SourceTypeMemory)

	if !p.HostConfig.Memory.MemoryDisabled {
		vm, err := mem.VirtualMemory()
		if err != nil {
			zap.L().Error("error on getting memory usage", zap.Error(err))
			return
		}
		msg.Payloads = append(msg.Payloads, p.getData("total", vm.Total, metricsML.MetricTypeNone))
		msg.Payloads = append(msg.Payloads, p.getData("used", vm.Used, metricsML.MetricTypeGauge))
		msg.Payloads = append(msg.Payloads, p.getData("used_percent", vm.UsedPercent, metricsML.MetricTypeGaugeFloat))
	}

	if !p.HostConfig.Memory.SwapMemoryDisabled {
		sm, err := mem.SwapMemory()
		if err != nil {
			zap.L().Error("error on getting swap memory usage", zap.Error(err))
			return
		}
		msg.Payloads = append(msg.Payloads, p.getData("swap_total", sm.Total, metricsML.MetricTypeNone))
		msg.Payloads = append(msg.Payloads, p.getData("swap_used", sm.Used, metricsML.MetricTypeGauge))
		msg.Payloads = append(msg.Payloads, p.getData("swap_used_percent", sm.UsedPercent, metricsML.MetricTypeGaugeFloat))
	}

	err = p.postMsg(&msg)
	if err != nil {
		zap.L().Error("error on posting msg", zap.Error(err))
		return
	}
}

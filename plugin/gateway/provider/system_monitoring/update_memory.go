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

	// memory config
	cfg := p.HostConfig.Memory

	// data message
	msg := p.getMsg(config.SourceTypeMemory)

	if !cfg.MemoryDisabled {
		vm, err := mem.VirtualMemory()
		if err != nil {
			zap.L().Error("error on getting memory usage", zap.Error(err))
			return
		}
		msg.Payloads = append(msg.Payloads, p.getData("total", getValueByUnit(vm.Total, cfg.Unit), metricsML.MetricTypeNone))
		msg.Payloads = append(msg.Payloads, p.getData("used", getValueByUnit(vm.Used, cfg.Unit), metricsML.MetricTypeGauge))
		msg.Payloads = append(msg.Payloads, p.getData("used_percent", vm.UsedPercent, metricsML.MetricTypeGaugeFloat))
	}

	if !cfg.SwapMemoryDisabled {
		sm, err := mem.SwapMemory()
		if err != nil {
			zap.L().Error("error on getting swap memory usage", zap.Error(err))
			return
		}
		msg.Payloads = append(msg.Payloads, p.getData("swap_total", getValueByUnit(sm.Total, cfg.Unit), metricsML.MetricTypeNone))
		msg.Payloads = append(msg.Payloads, p.getData("swap_used", getValueByUnit(sm.Used, cfg.Unit), metricsML.MetricTypeGauge))
		msg.Payloads = append(msg.Payloads, p.getData("swap_used_percent", sm.UsedPercent, metricsML.MetricTypeGaugeFloat))
	}

	err = p.postMsg(&msg)
	if err != nil {
		zap.L().Error("error on posting msg", zap.Error(err))
		return
	}
}

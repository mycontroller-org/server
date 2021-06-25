package systemmonitoring

import (
	"fmt"
	"time"

	coreScheduler "github.com/mycontroller-org/server/v2/pkg/service/core_scheduler"
	"github.com/mycontroller-org/server/v2/plugin/gateway/provider/system_monitoring/config"
	"go.uber.org/zap"
)

const (
	defaultInterval          = "1m"
	NodeDetailMetricInterval = "5m"
)

func (p *Provider) scheduleMonitoring() error {
	cfg := p.HostConfig

	// present node details every 5 minutes once
	p.schedule("node-info", NodeDetailMetricInterval, p.updateNodeDetails)

	// memory schedule
	if !cfg.Memory.MemoryDisabled || !cfg.Memory.SwapDisabled {
		p.schedule(config.SourceTypeMemory, getInterval(cfg.Memory.Interval), p.updateMemory)
	}

	// temperature schedule
	if !cfg.Temperature.DisabledAll {
		p.schedule(config.SourceTypeTemperature, getInterval(cfg.Temperature.Interval), p.updateTemperature)
	}

	// cpu schedule
	if !cfg.CPU.CPUDisabled || !cfg.CPU.PerCPUDisabled {
		p.schedule(config.SourceTypeCPU, getInterval(cfg.CPU.Interval), p.updateCPU)
	}

	// disk schedule
	if !cfg.Disk.Disabled && len(cfg.Disk.Data) > 0 {
		enabled := false
		for _, disk := range cfg.Disk.Data {
			if !disk.Disabled {
				enabled = true
				break
			}
		}
		if enabled {
			p.schedule(config.SourceTypeDisk, getInterval(cfg.Disk.Interval), p.updateDisk)
		}
	}

	// process schedule
	if !cfg.Process.Disabled && len(cfg.Process.Data) > 0 {
		enabled := false
		for _, process := range cfg.Process.Data {
			if !process.Disabled {
				enabled = true
				break
			}
		}
		if enabled {
			p.schedule(config.SourceTypeProcess, getInterval(cfg.Process.Interval), p.updateProcess)
		}
	}

	return nil
}

func getInterval(interval string) string {
	if interval == "" {
		return defaultInterval
	}
	_, err := time.ParseDuration(interval)
	if err != nil {
		return defaultInterval
	}
	return interval
}

func (p *Provider) getScheduleID(resourceID string) string {
	return fmt.Sprintf("%s_%s_%s", schedulePrefix, p.GatewayConfig.ID, resourceID)
}

func (p *Provider) unloadAll() {
	coreScheduler.SVC.RemoveWithPrefix(fmt.Sprintf("%s_%s", schedulePrefix, p.GatewayConfig.ID))
}

func (p *Provider) schedule(resourceID, interval string, triggerFunc func()) {
	p.unschedule(resourceID) // removes the existing schedule, if any
	schedulerID := p.getScheduleID(resourceID)
	cronSpec := fmt.Sprintf("@every %s", interval)
	err := coreScheduler.SVC.AddFunc(schedulerID, cronSpec, triggerFunc)
	if err != nil {
		zap.L().Error("error on adding schedule", zap.Error(err))
	}
	zap.L().Debug("added a schedule", zap.String("resourceID", resourceID), zap.String("schedulerID", schedulerID), zap.String("interval", interval))
}

func (p *Provider) unschedule(resourceID string) {
	schedulerID := p.getScheduleID(resourceID)
	coreScheduler.SVC.RemoveFunc(schedulerID)
	zap.L().Debug("removed a schedule", zap.String("resourceID", resourceID), zap.String("schedulerID", schedulerID))
}

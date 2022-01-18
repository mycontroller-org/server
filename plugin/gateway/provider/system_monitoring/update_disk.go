package systemmonitoring

import (
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	"github.com/shirou/gopsutil/v3/disk"
	"go.uber.org/zap"
)

func (p *Provider) updateDisk() {
	for sourceID, dataCFG := range p.HostConfig.Disk.Data {
		if !dataCFG.Disabled {
			stat, err := disk.Usage(dataCFG.Path)
			if err != nil {
				zap.L().Error("error on getting disk stat", zap.String("path", dataCFG.Path), zap.Error(err))
				continue
			}

			// presentation message
			sourceName := dataCFG.Name
			if sourceName == "" {
				sourceName = dataCFG.Path
			}
			presentMsg := p.getSourcePresentationMsg(sourceID, sourceName)
			othersData := presentMsg.Payloads[0]
			othersData.Others.Set("fstype", stat.Fstype, nil)
			othersData.Others.Set("inodes_total", stat.InodesTotal, nil)
			othersData.Others.Set("disk_size", getValueByUnit(stat.Total, dataCFG.Unit), nil)
			othersData.Others.Set("disk_size_unit", dataCFG.Unit, nil)
			presentMsg.Payloads[0] = othersData

			err = p.postMsg(&presentMsg)
			if err != nil {
				zap.L().Error("error on posting msg", zap.Error(err))
				return
			}

			// send data
			msg := p.getMsg(sourceID)
			inodeData := p.getData("inodes_used_percent", stat.InodesUsedPercent, metricTY.MetricTypeGaugeFloat, metricTY.UnitPercent, true)
			inodeData.Others.Set("used", stat.InodesUsed, nil)
			inodeData.Others.Set("total", stat.InodesTotal, nil)
			msg.Payloads = append(msg.Payloads, inodeData)

			usedData := p.getData("used_percent", stat.UsedPercent, metricTY.MetricTypeGaugeFloat, metricTY.UnitPercent, true)
			usedData.Others.Set("used", getValueByUnit(stat.Used, dataCFG.Unit), nil)
			usedData.Others.Set("total", getValueByUnit(stat.Total, dataCFG.Unit), nil)
			msg.Payloads = append(msg.Payloads, usedData)

			err = p.postMsg(&msg)
			if err != nil {
				zap.L().Error("error on posting msg", zap.Error(err))
				return
			}

		}
	}

}

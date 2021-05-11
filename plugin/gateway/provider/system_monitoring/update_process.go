package systemmonitoring

import (
	"fmt"
	"strings"

	"github.com/mycontroller-org/backend/v2/pkg/utils/convertor"
	metricsML "github.com/mycontroller-org/backend/v2/plugin/metrics"
	"github.com/shirou/gopsutil/v3/process"
	"go.uber.org/zap"
)

func (p *Provider) updateProcess() {
	procs, err := process.Processes()
	if err != nil {
		zap.L().Error("error on getting process list", zap.Error(err))
		return
	}

	for sourceID, dataCFG := range p.HostConfig.Process.Data {
		if dataCFG.Disabled {
			continue
		}
		if len(dataCFG.Filter) == 0 {
			continue
		}

		for _, proc := range procs {

			// verify filter
			matching := true
			for key, expectedValue := range dataCFG.Filter {
				actualValue := ""
				var err error
				switch strings.ToLower(key) {
				case "pid":
					actualValue = convertor.ToString(proc.Pid)

				case "cmdline":
					value, e := proc.Cmdline()
					actualValue = value
					err = e

				case "cwd":
					value, e := proc.Cwd()
					actualValue = value
					err = e

				case "exe":
					value, e := proc.Exe()
					actualValue = value
					err = e

				case "name":
					value, e := proc.Name()
					actualValue = value
					err = e

				case "nice":
					value, e := proc.Nice()
					actualValue = convertor.ToString(value)
					err = e

				case "ppid":
					value, e := proc.Ppid()
					actualValue = convertor.ToString(value)
					err = e

				case "username":
					value, e := proc.Username()
					actualValue = value
					err = e

				default:
					actualValue = fmt.Sprintf("%s not found", key)
				}

				if err != nil {
					if !strings.Contains(err.Error(), "no such file or directory") {
						zap.L().Error("error on collecting process data", zap.Error(err))
					}
					matching = false
					break
				}

				if expectedValue != actualValue {
					matching = false
					break
				}

			}

			if !matching {
				continue
			}

			// presentation message
			sourceName := dataCFG.Name
			if sourceName == "" {
				sourceName = sourceID
			}
			presentMsg := p.getSourcePresentationMsg(sourceID, sourceName)
			othersData := presentMsg.Payloads[0]

			othersData.Others.Set("pid", proc.Pid, nil)
			othersData.Labels.Set("pid", convertor.ToString(proc.Pid))

			var value interface{}

			value, err := proc.Cmdline()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set("cmdLine", value, nil)

			value, err = proc.Cwd()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set("cwd", value, nil)

			value, err = proc.Exe()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set("exe", value, nil)

			value, err = proc.Name()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set("name", value, nil)

			value, err = proc.Nice()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set("nice", value, nil)

			value, err = proc.Ppid()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set("ppid", value, nil)

			value, err = proc.Username()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set("username", value, nil)

			value, err = proc.Gids()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set("gids", value, nil)

			value, err = proc.Uids()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set("uids", value, nil)

			presentMsg.Payloads[0] = othersData

			err = p.postMsg(&presentMsg)
			if err != nil {
				zap.L().Error("error on posting msg", zap.Error(err))
				return
			}

			// update usage details
			msg := p.getMsg(sourceID)

			value, err = proc.CPUPercent()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			msg.Payloads = append(msg.Payloads, p.getData("cpu_percent", value, metricsML.MetricTypeGaugeFloat))

			value, err = proc.MemoryPercent()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			msg.Payloads = append(msg.Payloads, p.getData("memory_percent", value, metricsML.MetricTypeGaugeFloat))

			memInfo, err := proc.MemoryInfo()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			msg.Payloads = append(msg.Payloads, p.getData("rss", getValueByUnit(memInfo.RSS, dataCFG.Unit), metricsML.MetricTypeGauge))
			msg.Payloads = append(msg.Payloads, p.getData("vms", getValueByUnit(memInfo.VMS, dataCFG.Unit), metricsML.MetricTypeNone))
			msg.Payloads = append(msg.Payloads, p.getData("swap", getValueByUnit(memInfo.Swap, dataCFG.Unit), metricsML.MetricTypeNone))
			msg.Payloads = append(msg.Payloads, p.getData("stack", memInfo.Stack, metricsML.MetricTypeNone))
			msg.Payloads = append(msg.Payloads, p.getData("locked", memInfo.Locked, metricsML.MetricTypeNone))
			msg.Payloads = append(msg.Payloads, p.getData("data", memInfo.Data, metricsML.MetricTypeNone))

			err = p.postMsg(&msg)
			if err != nil {
				zap.L().Error("error on posting msg", zap.Error(err))
				return
			}

		}
	}

}

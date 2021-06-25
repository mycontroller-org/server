package systemmonitoring

import (
	"fmt"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/server/v2/plugin/gateway/provider/system_monitoring/config"
	metricsML "github.com/mycontroller-org/server/v2/plugin/metrics"
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
				case config.ProcessFieldPid:
					actualValue = convertor.ToString(proc.Pid)

				case config.ProcessFieldCmdLine:
					value, e := proc.Cmdline()
					actualValue = value
					err = e

				case config.ProcessFieldCwd:
					value, e := proc.Cwd()
					actualValue = value
					err = e

				case config.ProcessFieldEXE:
					value, e := proc.Exe()
					actualValue = value
					err = e

				case config.ProcessFieldName:
					value, e := proc.Name()
					actualValue = value
					err = e

				case config.ProcessFieldNice:
					value, e := proc.Nice()
					actualValue = convertor.ToString(value)
					err = e

				case config.ProcessFieldPPid:
					value, e := proc.Ppid()
					actualValue = convertor.ToString(value)
					err = e

				case config.ProcessFieldUsername:
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

			othersData.Others.Set(config.ProcessFieldPid, proc.Pid, nil)
			othersData.Labels.Set(config.ProcessFieldPid, convertor.ToString(proc.Pid))

			var value interface{}

			value, err := proc.Cmdline()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set(config.ProcessFieldCmdLine, value, nil)

			value, err = proc.Cwd()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set(config.ProcessFieldCwd, value, nil)

			value, err = proc.Exe()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set(config.ProcessFieldEXE, value, nil)

			value, err = proc.Name()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set(config.ProcessFieldName, value, nil)

			value, err = proc.Nice()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set(config.ProcessFieldNice, value, nil)

			value, err = proc.Ppid()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set(config.ProcessFieldPPid, value, nil)

			value, err = proc.Username()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set(config.ProcessFieldUsername, value, nil)

			value, err = proc.Gids()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set(config.ProcessFieldGids, value, nil)

			value, err = proc.Uids()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			othersData.Others.Set(config.ProcessFieldUids, value, nil)

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
			msg.Payloads = append(msg.Payloads, p.getData(config.ProcessFieldCpuPercent, value, metricsML.MetricTypeGaugeFloat, metricsML.UnitPercent, true))

			value, err = proc.MemoryPercent()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			msg.Payloads = append(msg.Payloads, p.getData(config.ProcessFieldMemoryPercent, value, metricsML.MetricTypeGaugeFloat, metricsML.UnitPercent, true))

			memInfo, err := proc.MemoryInfo()
			if err != nil {
				zap.L().Error("error on collecting process data", zap.Error(err))
				continue
			}
			msg.Payloads = append(msg.Payloads, p.getData(config.ProcessFieldRSS, getValueByUnit(memInfo.RSS, dataCFG.Unit), metricsML.MetricTypeGaugeFloat, dataCFG.Unit, true))
			msg.Payloads = append(msg.Payloads, p.getData(config.ProcessFieldVMS, getValueByUnit(memInfo.VMS, dataCFG.Unit), metricsML.MetricTypeNone, dataCFG.Unit, true))
			msg.Payloads = append(msg.Payloads, p.getData(config.ProcessFieldSwap, getValueByUnit(memInfo.Swap, dataCFG.Unit), metricsML.MetricTypeNone, dataCFG.Unit, true))
			msg.Payloads = append(msg.Payloads, p.getData(config.ProcessFieldStack, memInfo.Stack, metricsML.MetricTypeNone, metricsML.UnitNone, true))
			msg.Payloads = append(msg.Payloads, p.getData(config.ProcessFieldLocked, memInfo.Locked, metricsML.MetricTypeNone, metricsML.UnitNone, true))
			msg.Payloads = append(msg.Payloads, p.getData(config.ProcessFieldData, memInfo.Data, metricsML.MetricTypeNone, metricsML.UnitNone, true))

			err = p.postMsg(&msg)
			if err != nil {
				zap.L().Error("error on posting msg", zap.Error(err))
				return
			}

		}
	}

}

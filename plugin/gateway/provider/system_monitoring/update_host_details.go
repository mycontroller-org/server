package systemmonitoring

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/version"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	"github.com/shirou/gopsutil/v3/host"
	"go.uber.org/zap"
)

func (p *Provider) HostID() (string, error) {
	return host.HostID()
}

func (p *Provider) updateNodeDetails() {
	info, err := host.Info()
	if err != nil {
		zap.L().Error("error on getting node details", zap.Error(err))
		return
	}

	msg := msgTY.NewMessage(true)
	msg.GatewayID = p.GatewayConfig.ID
	msg.NodeID = p.NodeID
	msg.Type = msgTY.TypePresentation
	msg.Timestamp = time.Now()

	// struct to map
	infoMap := utils.StructToMap(&info)

	for name, value := range infoMap {
		msg.Payloads = append(msg.Payloads, p.getData(name, value, metricTY.MetricTypeNone, metricTY.UnitNone, false))
	}

	// update gateway version, gitcommit and build date
	gwVersion := version.Get()
	msg.Payloads = append(msg.Payloads, p.getData("gateway_version", gwVersion.Version, metricTY.MetricTypeNone, metricTY.UnitNone, false))
	msg.Payloads = append(msg.Payloads, p.getData("gateway_git_commit", gwVersion.GitCommit, metricTY.MetricTypeNone, metricTY.UnitNone, false))
	msg.Payloads = append(msg.Payloads, p.getData("gateway_build_date", gwVersion.BuildDate, metricTY.MetricTypeNone, metricTY.UnitNone, false))
	msg.Payloads = append(msg.Payloads, p.getData("gateway_platform", gwVersion.Platform, metricTY.MetricTypeNone, metricTY.UnitNone, false))
	msg.Payloads = append(msg.Payloads, p.getData("gateway_arch", gwVersion.Arch, metricTY.MetricTypeNone, metricTY.UnitNone, false))
	msg.Payloads = append(msg.Payloads, p.getData("gateway_golang_version", gwVersion.GoLang, metricTY.MetricTypeNone, metricTY.UnitNone, false))

	// include version details
	// library_version
	data := p.getData("name", info.Hostname, metricTY.MetricTypeNone, metricTY.UnitNone, false)
	data.Labels.Set(types.LabelNodeVersion, fmt.Sprintf("%s_%s", info.PlatformFamily, info.PlatformVersion))
	data.Labels.Set(types.LabelNodeLibraryVersion, info.KernelVersion)
	data.Labels.Set("arch", info.KernelArch)
	data.Labels.Set("os", info.OS)

	msg.Payloads = append(msg.Payloads, data)

	err = p.postMsg(&msg)
	if err != nil {
		// return err
		zap.L().Error("error on posting msg", zap.Error(err))
		return
	}
}

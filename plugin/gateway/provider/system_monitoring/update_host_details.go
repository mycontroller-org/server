package systemmonitoring

import (
	"fmt"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/model"
	msgML "github.com/mycontroller-org/server/v2/pkg/model/message"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/version"
	metricType "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
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

	msg := msgML.NewMessage(true)
	msg.GatewayID = p.GatewayConfig.ID
	msg.NodeID = p.NodeID
	msg.Type = msgML.TypePresentation
	msg.Timestamp = time.Now()

	// struct to map
	infoMap := utils.StructToMap(&info)

	for name, value := range infoMap {
		msg.Payloads = append(msg.Payloads, p.getData(name, value, metricType.MetricTypeNone, metricType.UnitNone, false))
	}

	// update gateway version, gitcommit and build date
	gwVersion := version.Get()
	msg.Payloads = append(msg.Payloads, p.getData("gateway_version", gwVersion.Version, metricType.MetricTypeNone, metricType.UnitNone, false))
	msg.Payloads = append(msg.Payloads, p.getData("gateway_git_commit", gwVersion.GitCommit, metricType.MetricTypeNone, metricType.UnitNone, false))
	msg.Payloads = append(msg.Payloads, p.getData("gateway_build_date", gwVersion.BuildDate, metricType.MetricTypeNone, metricType.UnitNone, false))
	msg.Payloads = append(msg.Payloads, p.getData("gateway_platform", gwVersion.Platform, metricType.MetricTypeNone, metricType.UnitNone, false))
	msg.Payloads = append(msg.Payloads, p.getData("gateway_arch", gwVersion.Arch, metricType.MetricTypeNone, metricType.UnitNone, false))
	msg.Payloads = append(msg.Payloads, p.getData("gateway_golang_version", gwVersion.GoLang, metricType.MetricTypeNone, metricType.UnitNone, false))

	// include version details
	// library_version
	data := p.getData("name", info.Hostname, metricType.MetricTypeNone, metricType.UnitNone, false)
	data.Labels.Set(model.LabelNodeVersion, fmt.Sprintf("%s_%s", info.PlatformFamily, info.PlatformVersion))
	data.Labels.Set(model.LabelNodeLibraryVersion, info.KernelVersion)
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

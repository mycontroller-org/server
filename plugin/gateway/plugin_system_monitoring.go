package gateway

import (
	systemMonitoring "github.com/mycontroller-org/server/v2/plugin/gateway/provider/system_monitoring"
)

func init() {
	Register(systemMonitoring.PluginSystemMonitoring, systemMonitoring.NewPluginSystemMonitoring)
}

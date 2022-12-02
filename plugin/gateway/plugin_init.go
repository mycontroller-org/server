package gateway

import (
	esphome "github.com/mycontroller-org/server/v2/plugin/gateway/provider/esphome"
	generic "github.com/mycontroller-org/server/v2/plugin/gateway/provider/generic"
	mysensorsV2 "github.com/mycontroller-org/server/v2/plugin/gateway/provider/mysensors_v2"
	philipsHue "github.com/mycontroller-org/server/v2/plugin/gateway/provider/philipshue"
	systemMonitoring "github.com/mycontroller-org/server/v2/plugin/gateway/provider/system_monitoring"
	"github.com/mycontroller-org/server/v2/plugin/gateway/provider/tasmota"
)

func init() {
	Register(esphome.PluginEspHome, esphome.NewPluginEspHome)
	Register(generic.PluginGeneric, generic.NewPluginGeneric)
	Register(mysensorsV2.PluginMySensorsV2, mysensorsV2.NewPluginMySensorsV2)
	Register(philipsHue.PluginPhilipsHue, philipsHue.NewPluginPhilipsHue)
	Register(systemMonitoring.PluginSystemMonitoring, systemMonitoring.NewPluginSystemMonitoring)
	Register(tasmota.PluginTasmota, tasmota.NewPluginTasmota)
}

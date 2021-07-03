package gateway

import (
	"github.com/mycontroller-org/server/v2/plugin/gateway/provider/tasmota"
)

func init() {
	Register(tasmota.PluginTasmota, tasmota.NewPluginTasmota)
}

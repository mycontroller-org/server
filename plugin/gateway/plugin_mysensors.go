package gateway

import (
	mysensorsV2 "github.com/mycontroller-org/server/v2/plugin/gateway/provider/mysensors_v2"
)

func init() {
	Register(mysensorsV2.PluginMySensorsV2, mysensorsV2.NewPluginMySensorsV2)
}

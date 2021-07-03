package gateway

import (
	esphome "github.com/mycontroller-org/server/v2/plugin/gateway/provider/esphome"
)

func init() {
	Register(esphome.PluginEspHome, esphome.NewPluginEspHome)
}

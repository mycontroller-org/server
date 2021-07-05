package bus

import (
	embedded "github.com/mycontroller-org/server/v2/plugin/bus/embedded"
)

func init() {
	Register(embedded.PluginEmbedded, embedded.NewClient)
}

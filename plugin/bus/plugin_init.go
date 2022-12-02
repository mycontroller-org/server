package bus

import (
	embedded "github.com/mycontroller-org/server/v2/plugin/bus/embedded"
	natsIO "github.com/mycontroller-org/server/v2/plugin/bus/natsio"
)

func init() {
	Register(embedded.PluginEmbedded, embedded.NewClient)
	Register(natsIO.PluginNATSIO, natsIO.NewClient)
}

package bus

import (
	natsIO "github.com/mycontroller-org/server/v2/plugin/bus/natsio"
)

func init() {
	Register(natsIO.PluginNATSIO, natsIO.NewClient)
}

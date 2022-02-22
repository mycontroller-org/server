package gateway

import (
	generic "github.com/mycontroller-org/server/v2/plugin/gateway/provider/generic"
)

func init() {
	Register(generic.PluginGeneric, generic.NewPluginGeneric)
}

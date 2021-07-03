package gateway

import (
	philipshue "github.com/mycontroller-org/server/v2/plugin/gateway/provider/philipshue"
)

func init() {
	Register(philipshue.PluginPhilipsHue, philipshue.NewPluginPhilipsHue)
}

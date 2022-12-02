package metric

import (
	influxdbV2 "github.com/mycontroller-org/server/v2/plugin/database/metric/influxdb_v2"
	voiddb "github.com/mycontroller-org/server/v2/plugin/database/metric/voiddb"
)

func init() {
	Register(voiddb.PluginVoidDB, voiddb.NewClient)
	Register(influxdbV2.PluginInfluxdbV2, influxdbV2.NewClient)
}

package metric

import (
	voiddb "github.com/mycontroller-org/server/v2/plugin/database/metric/voiddb"
)

func init() {
	Register(voiddb.PluginVoidDB, voiddb.NewClient)
}

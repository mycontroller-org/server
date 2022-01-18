package store

import (
	cfgTY "github.com/mycontroller-org/server/v2/pkg/types/config"
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

var (
	CFG     *cfgTY.Config
	STORAGE storageTY.Plugin
	METRIC  metricTY.Plugin
)

func InitConfig(config *cfgTY.Config) {
	CFG = config
}

func InitStorage(storage storageTY.Plugin) {
	STORAGE = storage
}

func InitMetric(metric metricTY.Plugin) {
	METRIC = metric
}

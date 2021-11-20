package store

import (
	cfgML "github.com/mycontroller-org/server/v2/pkg/model/config"
	metricType "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

var (
	CFG     *cfgML.Config
	STORAGE stgType.Plugin
	METRIC  metricType.Plugin
)

func InitConfig(config *cfgML.Config) {
	CFG = config
}

func InitStorage(storage stgType.Plugin) {
	STORAGE = storage
}

func InitMetric(metric metricType.Plugin) {
	METRIC = metric
}

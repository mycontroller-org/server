package service

import (
	"errors"
	"fmt"

	mtsml "github.com/mycontroller-org/backend/v2/plugin/metrics"
	influx "github.com/mycontroller-org/backend/v2/plugin/metrics/influxdb_v2"
	"github.com/mycontroller-org/backend/v2/plugin/metrics/voiddb"
)

// InitMetricsDatabase metrics database
func InitMetricsDatabase(config map[string]interface{}) (mtsml.Client, error) {
	dbType, available := config["type"]
	if available {
		switch dbType {
		case mtsml.TypeInfluxdbV2:
			return influx.NewClient(config)
		case mtsml.TypeVoidDB:
			return voiddb.NewClient(config)
		default:
			return nil, fmt.Errorf("Specified database type not implemented. %s", dbType)
		}
	}
	return nil, errors.New("'type' field should be added on the database config")
}

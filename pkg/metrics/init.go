package metrics

import (
	"errors"
	"fmt"

	mtsml "github.com/mycontroller-org/backend/v2/pkg/model/metrics"
	influx "github.com/mycontroller-org/backend/v2/plugin/metrics/influxdb_v2"
	"github.com/mycontroller-org/backend/v2/plugin/metrics/voiddb"
)

// Init metrics database
func Init(config map[string]interface{}) (mtsml.Client, error) {
	dbType, available := config["type"]
	if available {
		switch dbType {
		case mtsml.DBTypeInfluxdbV2:
			return influx.NewClient(config)
		case mtsml.DBTypeVoidDB:
			return voiddb.NewClient(config)
		default:
			return nil, fmt.Errorf("Specified database type not implemented. %s", dbType)
		}
	}
	return nil, errors.New("'type' field should be added on the database config")
}

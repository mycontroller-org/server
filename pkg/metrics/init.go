package metrics

import (
	"errors"
	"fmt"

	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	mtrml "github.com/mycontroller-org/backend/v2/pkg/model/metric"
	influx "github.com/mycontroller-org/backend/v2/plugin/metrics/influxdb_v2"
	"github.com/mycontroller-org/backend/v2/plugin/metrics/voiddb"
)

// Metrics database types
const (
	TypeInfluxdbV2 = "influxdb_v2"
	TypeVoidDB     = "void_db"
)

// Client interface
type Client interface {
	Close() error
	Ping() error
	Write(field *fml.Field) error
	WriteBlocking(field *fml.Field) error
	Query(queryConfig *mtrml.QueryConfig) (map[string][]mtrml.Data, error)
	//Query(metricType, interval string)
}

// Init metrics database
func Init(config map[string]interface{}) (Client, error) {
	dbType, available := config["type"]
	if available {
		switch dbType {
		case TypeInfluxdbV2:
			return influx.NewClient(config)
		case TypeVoidDB:
			return voiddb.NewClient(config)
		default:
			return nil, fmt.Errorf("Specified database type not implemented. %s", dbType)
		}
	}
	return nil, errors.New("'type' field should be added on the database config")
}

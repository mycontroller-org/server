package metrics

import (
	"errors"
	"fmt"

	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	influx "github.com/mycontroller-org/mycontroller-v2/plugin/storage/metrics/influxdb_v2"
	"github.com/mycontroller-org/mycontroller-v2/plugin/storage/metrics/voiddb"
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
	Write(variable *ml.SensorField) error
	WriteBlocking(variable *ml.SensorField) error
}

// Init metrics database
func Init(config map[string]interface{}) (Client, error) {
	dbType, available := config["type"]
	if available {
		switch dbType {
		case TypeInfluxdbV2:
			c, err := influx.NewClient(config)
			if err != nil {
				return nil, err
			}
			var cl Client = c
			return cl, nil
		case TypeVoidDB:
			c, err := voiddb.NewClient(config)
			var cl Client = c
			return cl, err
		default:
			return nil, fmt.Errorf("Specified database type not implemented. %s", dbType)
		}
	}
	return nil, errors.New("'type' field should be added on the database config")
}

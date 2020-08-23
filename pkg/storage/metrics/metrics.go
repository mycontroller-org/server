package metrics

import (
	"errors"
	"fmt"

	sml "github.com/mycontroller-org/backend/pkg/model/sensor"
	influx "github.com/mycontroller-org/backend/plugin/storage/metrics/influxdb_v2"
	"github.com/mycontroller-org/backend/plugin/storage/metrics/voiddb"
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
	Write(variable *sml.SensorField) error
	WriteBlocking(variable *sml.SensorField) error
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

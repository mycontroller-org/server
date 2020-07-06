package metrics

import (
	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
	influx "github.com/mycontroller-org/mycontroller-v2/plugin/storage/metrics/influxdb_v2"
)

// Client interface
type Client interface {
	Close() error
	Ping() error
	Write(variable *ml.SensorField) error
	WriteBlocking(variable *ml.SensorField) error
}

// Init storage
func Init(config map[string]interface{}) (*Client, error) {
	c, err := influx.NewClient(config)
	if err != nil {
		return nil, err
	}
	var cl Client = c
	return &cl, nil
}

package metrics

import (
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
)

// Metrics database types
const (
	DBTypeInfluxdbV2 = "influxdb_v2"
	DBTypeVoidDB     = "void_db"
)

// Client interface
type Client interface {
	Close() error
	Ping() error
	Write(field *fml.Field) error
	WriteBlocking(field *fml.Field) error
	Query(queryConfig *QueryConfig) (map[string][]Data, error)
}

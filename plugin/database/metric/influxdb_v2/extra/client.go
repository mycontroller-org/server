package query

import (
	metricType "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
)

// contants
const (
	DefaultWindow = "5m"
	DefaultStart  = "-1h"

	FieldValue     = "value"
	FieldLatitude  = "latitude"
	FieldLongitude = "longitude"
	FieldAltitude  = "altitude"
)

// QueryAPI interface
type QueryAPI interface {
	ExecuteQuery(queryConfig *metricType.Query, measurement string) ([]metricType.ResponseData, error)
}

// AdminAPI interface
type AdminAPI interface {
	CreateBucket() error
}

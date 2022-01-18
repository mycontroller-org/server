package query

import (
	metricTY "github.com/mycontroller-org/server/v2/plugin/database/metric/types"
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
	ExecuteQuery(queryConfig *metricTY.Query, measurement string) ([]metricTY.ResponseData, error)
}

// AdminAPI interface
type AdminAPI interface {
	CreateBucket() error
}

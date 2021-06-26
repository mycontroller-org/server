package query

import (
	metricsML "github.com/mycontroller-org/server/v2/plugin/database/metrics"
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
	ExecuteQuery(queryConfig *metricsML.Query, measurement string) ([]metricsML.ResponseData, error)
}

// AdminAPI interface
type AdminAPI interface {
	CreateBucket() error
}

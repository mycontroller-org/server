package query

import (
	mtrml "github.com/mycontroller-org/backend/v2/plugin/metrics"
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
	ExecuteQuery(queryConfig *mtrml.Query, measurement string) ([]mtrml.ResponseData, error)
}

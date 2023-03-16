package api

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

func ApplyDeviceFilter(labelFilters cmap.CustomStringMap, filters []storageTY.Filter) []storageTY.Filter {
	if filters == nil {
		filters = make([]storageTY.Filter, 0)
	}

	// create labels filter
	labelsFilter := make([]storageTY.Filter, 0)
	for key, value := range labelFilters {
		labelsFilter = append(labelsFilter, storageTY.Filter{
			Key:      fmt.Sprintf("labels.%s", key),
			Operator: storageTY.OperatorEqual,
			Value:    value,
		})
	}

	return append(filters, labelsFilter...)
}

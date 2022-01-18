package dashboard

import (
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	dashboardTY "github.com/mycontroller-org/server/v2/pkg/types/dashboard"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]dashboardTY.Config, 0)
	return store.STORAGE.Find(types.EntityDashboard, &result, filters, pagination)
}

// Get returns an item
func Get(filters []storageTY.Filter) (*dashboardTY.Config, error) {
	result := &dashboardTY.Config{}
	err := store.STORAGE.FindOne(types.EntityDashboard, result, filters)
	return result, err
}

// Save an item
func Save(dashboard *dashboardTY.Config) error {
	if dashboard.ID == "" {
		dashboard.ID = utils.RandUUID()
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: dashboard.ID},
	}
	return store.STORAGE.Upsert(types.EntityDashboard, dashboard, filters)
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(types.EntityDashboard, filters)
}

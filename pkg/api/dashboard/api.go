package dashboard

import (
	"github.com/mycontroller-org/server/v2/pkg/model"
	dashboardML "github.com/mycontroller-org/server/v2/pkg/model/dashboard"
	"github.com/mycontroller-org/server/v2/pkg/store"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	stgType "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// List by filter and pagination
func List(filters []stgType.Filter, pagination *stgType.Pagination) (*stgType.Result, error) {
	result := make([]dashboardML.Config, 0)
	return store.STORAGE.Find(model.EntityDashboard, &result, filters, pagination)
}

// Get returns an item
func Get(filters []stgType.Filter) (*dashboardML.Config, error) {
	result := &dashboardML.Config{}
	err := store.STORAGE.FindOne(model.EntityDashboard, result, filters)
	return result, err
}

// Save an item
func Save(dashboard *dashboardML.Config) error {
	if dashboard.ID == "" {
		dashboard.ID = utils.RandUUID()
	}
	filters := []stgType.Filter{
		{Key: model.KeyID, Value: dashboard.ID},
	}
	return store.STORAGE.Upsert(model.EntityDashboard, dashboard, filters)
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []stgType.Filter{{Key: model.KeyID, Operator: stgType.OperatorIn, Value: IDs}}
	return store.STORAGE.Delete(model.EntityDashboard, filters)
}

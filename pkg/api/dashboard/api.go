package dashboard

import (
	"github.com/mycontroller-org/backend/v2/pkg/model"
	dashboardML "github.com/mycontroller-org/backend/v2/pkg/model/dashboard"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	stgML "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgML.Filter, pagination *stgML.Pagination) (*stgML.Result, error) {
	result := make([]dashboardML.Config, 0)
	return stg.SVC.Find(model.EntityDashboard, &result, filters, pagination)
}

// Get returns an item
func Get(filters []stgML.Filter) (*dashboardML.Config, error) {
	result := &dashboardML.Config{}
	err := stg.SVC.FindOne(model.EntityDashboard, result, filters)
	return result, err
}

// Save an item
func Save(dashboard *dashboardML.Config) error {
	if dashboard.ID == "" {
		dashboard.ID = utils.RandUUID()
	}
	filters := []stgML.Filter{
		{Key: model.KeyID, Value: dashboard.ID},
	}
	return stg.SVC.Upsert(model.EntityDashboard, dashboard, filters)
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []stgML.Filter{{Key: model.KeyID, Operator: stgML.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(model.EntityDashboard, filters)
}

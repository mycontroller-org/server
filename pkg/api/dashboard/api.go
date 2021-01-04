package dashboard

import (
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	dbml "github.com/mycontroller-org/backend/v2/pkg/model/dashboard"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]dbml.Config, 0)
	return stg.SVC.Find(ml.EntityDashboard, &result, filters, pagination)
}

// Get returns an item
func Get(filters []stgml.Filter) (*dbml.Config, error) {
	result := &dbml.Config{}
	err := stg.SVC.FindOne(ml.EntityDashboard, result, filters)
	return result, err
}

// Save an item
func Save(dashboard *dbml.Config) error {
	if dashboard.ID == "" {
		dashboard.ID = utils.RandUUID()
	}
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: dashboard.ID},
	}
	return stg.SVC.Upsert(ml.EntityDashboard, dashboard, filters)
}

// Delete items
func Delete(IDs []string) (int64, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(ml.EntityDashboard, filters)
}

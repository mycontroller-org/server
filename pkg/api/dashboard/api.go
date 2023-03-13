package dashboard

import (
	"context"
	"fmt"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	dashboardTY "github.com/mycontroller-org/server/v2/pkg/types/dashboard"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

type DashboardAPI struct {
	ctx     context.Context
	logger  *zap.Logger
	storage storageTY.Plugin
}

func New(ctx context.Context, logger *zap.Logger, storage storageTY.Plugin) *DashboardAPI {
	return &DashboardAPI{
		ctx:     ctx,
		logger:  logger.Named("dashboard_api"),
		storage: storage,
	}
}

// List by filter and pagination
func (d *DashboardAPI) List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]dashboardTY.Config, 0)
	return d.storage.Find(types.EntityDashboard, &result, filters, pagination)
}

// Get returns an item
func (d *DashboardAPI) Get(filters []storageTY.Filter) (*dashboardTY.Config, error) {
	result := &dashboardTY.Config{}
	err := d.storage.FindOne(types.EntityDashboard, result, filters)
	return result, err
}

// Save an item
func (d *DashboardAPI) Save(dashboard *dashboardTY.Config) error {
	if dashboard.ID == "" {
		dashboard.ID = utils.RandUUID()
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: dashboard.ID},
	}
	return d.storage.Upsert(types.EntityDashboard, dashboard, filters)
}

// Delete items
func (d *DashboardAPI) Delete(IDs []string) (int64, error) {
	filters := []storageTY.Filter{{Key: types.KeyID, Operator: storageTY.OperatorIn, Value: IDs}}
	return d.storage.Delete(types.EntityDashboard, filters)
}

func (d *DashboardAPI) Import(data interface{}) error {
	input, ok := data.(dashboardTY.Config)
	if !ok {
		return fmt.Errorf("invalid type:%T", data)
	}
	if input.ID == "" {
		input.ID = utils.RandUUID()
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: input.ID},
	}

	return d.storage.Upsert(types.EntityDashboard, &input, filters)
}

func (d *DashboardAPI) GetEntityInterface() interface{} {
	return dashboardTY.Config{}
}

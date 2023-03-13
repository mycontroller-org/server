package routes

import (
	"errors"
	"fmt"
	"net/http"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	dashboardTY "github.com/mycontroller-org/server/v2/pkg/types/dashboard"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterDashboardRoutes registers dashboard api
func (h *Routes) registerDashboardRoutes() {
	h.router.HandleFunc("/api/dashboard", h.listDashboards).Methods(http.MethodGet)
	h.router.HandleFunc("/api/dashboard/{id}", h.getDashboard).Methods(http.MethodGet)
	h.router.HandleFunc("/api/dashboard", h.updateDashboard).Methods(http.MethodPost)
	h.router.HandleFunc("/api/dashboard", h.deleteDashboards).Methods(http.MethodDelete)
}

func (h *Routes) listDashboards(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntityDashboard, &[]dashboardTY.Config{})
}

func (h *Routes) getDashboard(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityDashboard, &dashboardTY.Config{})
}

func (h *Routes) updateDashboard(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]storageTY.Filter) error {
		entity := d.(*dashboardTY.Config)
		if entity.ID == "" {
			return errors.New("id should not be an empty")
		}
		return nil
	}
	handlerUtils.SaveEntity(h.storage, w, r, types.EntityDashboard, &dashboardTY.Config{}, bwFunc)
}

func (h *Routes) deleteDashboards(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.Dashboard().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	dashboardAPI "github.com/mycontroller-org/server/v2/pkg/api/dashboard"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	dashboardTY "github.com/mycontroller-org/server/v2/pkg/types/dashboard"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// RegisterDashboardRoutes registers dashboard api
func RegisterDashboardRoutes(router *mux.Router) {
	router.HandleFunc("/api/dashboard", listDashboards).Methods(http.MethodGet)
	router.HandleFunc("/api/dashboard/{id}", getDashboard).Methods(http.MethodGet)
	router.HandleFunc("/api/dashboard", updateDashboard).Methods(http.MethodPost)
	router.HandleFunc("/api/dashboard", deleteDashboards).Methods(http.MethodDelete)
}

func listDashboards(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, types.EntityDashboard, &[]dashboardTY.Config{})
}

func getDashboard(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, types.EntityDashboard, &dashboardTY.Config{})
}

func updateDashboard(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]storageTY.Filter) error {
		entity := d.(*dashboardTY.Config)
		if entity.ID == "" {
			return errors.New("id should not be an empty")
		}
		return nil
	}
	handlerUtils.SaveEntity(w, r, types.EntityDashboard, &dashboardTY.Config{}, bwFunc)
}

func deleteDashboards(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := dashboardAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

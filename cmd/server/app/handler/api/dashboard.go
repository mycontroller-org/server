package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	dashboardAPI "github.com/mycontroller-org/server/v2/pkg/api/dashboard"
	"github.com/mycontroller-org/server/v2/pkg/model"
	dashboardML "github.com/mycontroller-org/server/v2/pkg/model/dashboard"
	stgML "github.com/mycontroller-org/server/v2/plugin/storage"
)

// RegisterDashboardRoutes registers dashboard api
func RegisterDashboardRoutes(router *mux.Router) {
	router.HandleFunc("/api/dashboard", listDashboards).Methods(http.MethodGet)
	router.HandleFunc("/api/dashboard/{id}", getDashboard).Methods(http.MethodGet)
	router.HandleFunc("/api/dashboard", updateDashboard).Methods(http.MethodPost)
	router.HandleFunc("/api/dashboard", deleteDashboards).Methods(http.MethodDelete)
}

func listDashboards(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, model.EntityDashboard, &[]dashboardML.Config{})
}

func getDashboard(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, model.EntityDashboard, &dashboardML.Config{})
}

func updateDashboard(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]stgML.Filter) error {
		entity := d.(*dashboardML.Config)
		if entity.ID == "" {
			return errors.New("id should not be an empty")
		}
		return nil
	}
	handlerUtils.SaveEntity(w, r, model.EntityDashboard, &dashboardML.Config{}, bwFunc)
}

func deleteDashboards(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
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

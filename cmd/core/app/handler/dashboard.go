package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	dashboardAPI "github.com/mycontroller-org/backend/v2/pkg/api/dashboard"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	dbml "github.com/mycontroller-org/backend/v2/pkg/model/dashboard"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func registerDashboardRoutes(router *mux.Router) {
	router.HandleFunc("/api/dashboard", listDashboards).Methods(http.MethodGet)
	router.HandleFunc("/api/dashboard/{id}", getDashboard).Methods(http.MethodGet)
	router.HandleFunc("/api/dashboard", updateDashboard).Methods(http.MethodPost)
	router.HandleFunc("/api/dashboard", deleteDashboards).Methods(http.MethodDelete)
}

func listDashboards(w http.ResponseWriter, r *http.Request) {
	FindMany(w, r, ml.EntityDashboard, &[]dbml.Config{})
}

func getDashboard(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, ml.EntityDashboard, &dbml.Config{})
}

func updateDashboard(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]stgml.Filter) error {
		entity := d.(*dbml.Config)
		if entity.ID == "" {
			return errors.New("ID should not be an empty")
		}
		return nil
	}
	SaveEntity(w, r, ml.EntityDashboard, &dbml.Config{}, bwFunc)
}

func deleteDashboards(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := dashboardAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("Deleted: %d", count), nil
		}
		return nil, errors.New("Supply id(s)")
	}
	UpdateData(w, r, &IDs, updateFn)
}

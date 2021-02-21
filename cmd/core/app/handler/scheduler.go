package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	schedulerAPI "github.com/mycontroller-org/backend/v2/pkg/api/scheduler"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	schedulerML "github.com/mycontroller-org/backend/v2/pkg/model/scheduler"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func registerSchedulerRoutes(router *mux.Router) {
	router.HandleFunc("/api/scheduler", listSchedule).Methods(http.MethodGet)
	router.HandleFunc("/api/scheduler/{id}", getSchedule).Methods(http.MethodGet)
	router.HandleFunc("/api/scheduler", updateSchedule).Methods(http.MethodPost)
	router.HandleFunc("/api/scheduler/enable", enableSchedule).Methods(http.MethodPost)
	router.HandleFunc("/api/scheduler/disable", disableSchedule).Methods(http.MethodPost)
	router.HandleFunc("/api/scheduler", deleteSchedule).Methods(http.MethodDelete)
}

func listSchedule(w http.ResponseWriter, r *http.Request) {
	FindMany(w, r, ml.EntityScheduler, &[]schedulerML.Config{})
}

func getSchedule(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, ml.EntityScheduler, &schedulerML.Config{})
}

func updateSchedule(w http.ResponseWriter, r *http.Request) {
	entity := &schedulerML.Config{}
	err := LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "ID should not be an empty", 400)
		return
	}
	err = schedulerAPI.SaveAndReload(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func deleteSchedule(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := schedulerAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("Deleted: %d", count), nil
		}
		return nil, errors.New("Supply id(s)")
	}
	UpdateData(w, r, &IDs, updateFn)
}

func enableSchedule(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := schedulerAPI.Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("Supply a task id")
	}
	UpdateData(w, r, &ids, updateFn)
}

func disableSchedule(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := schedulerAPI.Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("Supply a task id")
	}
	UpdateData(w, r, &ids, updateFn)
}

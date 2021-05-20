package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	scheduleAPI "github.com/mycontroller-org/backend/v2/pkg/api/schedule"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	schedulerML "github.com/mycontroller-org/backend/v2/pkg/model/schedule"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func registerSchedulerRoutes(router *mux.Router) {
	router.HandleFunc("/api/schedule", listSchedule).Methods(http.MethodGet)
	router.HandleFunc("/api/schedule/{id}", getSchedule).Methods(http.MethodGet)
	router.HandleFunc("/api/schedule", updateSchedule).Methods(http.MethodPost)
	router.HandleFunc("/api/schedule/enable", enableSchedule).Methods(http.MethodPost)
	router.HandleFunc("/api/schedule/disable", disableSchedule).Methods(http.MethodPost)
	router.HandleFunc("/api/schedule", deleteSchedule).Methods(http.MethodDelete)
}

func listSchedule(w http.ResponseWriter, r *http.Request) {
	FindMany(w, r, ml.EntitySchedule, &[]schedulerML.Config{})
}

func getSchedule(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, ml.EntitySchedule, &schedulerML.Config{})
}

func updateSchedule(w http.ResponseWriter, r *http.Request) {
	entity := &schedulerML.Config{}
	err := LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be an empty", 400)
		return
	}
	err = scheduleAPI.SaveAndReload(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func deleteSchedule(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := scheduleAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	UpdateData(w, r, &IDs, updateFn)
}

func enableSchedule(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := scheduleAPI.Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a task id")
	}
	UpdateData(w, r, &ids, updateFn)
}

func disableSchedule(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := scheduleAPI.Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a task id")
	}
	UpdateData(w, r, &ids, updateFn)
}

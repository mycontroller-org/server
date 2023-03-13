package routes

import (
	"errors"
	"fmt"
	"net/http"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterSchedulerRoutes registers schedule api
func (h *Routes) registerSchedulerRoutes() {
	h.router.HandleFunc("/api/schedule", h.listSchedule).Methods(http.MethodGet)
	h.router.HandleFunc("/api/schedule/{id}", h.getSchedule).Methods(http.MethodGet)
	h.router.HandleFunc("/api/schedule", h.updateSchedule).Methods(http.MethodPost)
	h.router.HandleFunc("/api/schedule/enable", h.enableSchedule).Methods(http.MethodPost)
	h.router.HandleFunc("/api/schedule/disable", h.disableSchedule).Methods(http.MethodPost)
	h.router.HandleFunc("/api/schedule", h.deleteSchedule).Methods(http.MethodDelete)
}

func (h *Routes) listSchedule(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntitySchedule, &[]schedulerTY.Config{})
}

func (h *Routes) getSchedule(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntitySchedule, &schedulerTY.Config{})
}

func (h *Routes) updateSchedule(w http.ResponseWriter, r *http.Request) {
	entity := &schedulerTY.Config{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be an empty", http.StatusBadRequest)
		return
	}
	err = h.api.Schedule().SaveAndReload(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.Schedule().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

func (h *Routes) enableSchedule(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.Schedule().Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a task id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) disableSchedule(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.Schedule().Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a task id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

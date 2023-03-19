package routes

import (
	"errors"
	"fmt"
	"net/http"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterTaskRoutes registers task api
func (h *Routes) registerTaskRoutes() {
	h.router.HandleFunc("/api/task", h.listTasks).Methods(http.MethodGet)
	h.router.HandleFunc("/api/task/{id}", h.getTask).Methods(http.MethodGet)
	h.router.HandleFunc("/api/task", h.updateTask).Methods(http.MethodPost)
	h.router.HandleFunc("/api/task/enable", h.enableTask).Methods(http.MethodPost)
	h.router.HandleFunc("/api/task/disable", h.disableTask).Methods(http.MethodPost)
	h.router.HandleFunc("/api/task", h.deleteTasks).Methods(http.MethodDelete)
}

func (h *Routes) listTasks(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntityTask, &[]taskTY.Config{})
}

func (h *Routes) getTask(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityTask, &taskTY.Config{})
}

func (h *Routes) updateTask(w http.ResponseWriter, r *http.Request) {
	entity := &taskTY.Config{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be an empty", http.StatusBadRequest)
		return
	}
	err = h.api.Task().SaveAndReload(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) deleteTasks(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.Task().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

func (h *Routes) enableTask(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.Task().Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a task id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) disableTask(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.Task().Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a task id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

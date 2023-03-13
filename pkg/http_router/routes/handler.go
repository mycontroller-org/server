package routes

import (
	"errors"
	"fmt"
	"net/http"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
)

// RegisterHandlerRoutes registers handler api
func (h *Routes) registerHandlerRoutes() {
	h.router.HandleFunc("/api/handler", h.listHandler).Methods(http.MethodGet)
	h.router.HandleFunc("/api/handler/{id}", h.getHandler).Methods(http.MethodGet)
	h.router.HandleFunc("/api/handler", h.updateHandler).Methods(http.MethodPost)
	h.router.HandleFunc("/api/handler/enable", h.enableHandler).Methods(http.MethodPost)
	h.router.HandleFunc("/api/handler/disable", h.disableHandler).Methods(http.MethodPost)
	h.router.HandleFunc("/api/handler/reload", h.reloadHandler).Methods(http.MethodPost)
	h.router.HandleFunc("/api/handler", h.deleteHandler).Methods(http.MethodDelete)
}

func (h *Routes) listHandler(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntityHandler, &[]handlerTY.Config{})
}

func (h *Routes) getHandler(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityHandler, &handlerTY.Config{})
}

func (h *Routes) updateHandler(w http.ResponseWriter, r *http.Request) {
	entity := &handlerTY.Config{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be an empty", http.StatusBadRequest)
		return
	}
	err = h.api.Handler().SaveAndReload(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) enableHandler(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.Handler().Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a handler id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) disableHandler(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.Handler().Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a handler id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) reloadHandler(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.Handler().Reload(ids)
			if err != nil {
				return nil, err
			}
			return "Reloaded", nil
		}
		return nil, errors.New("supply a handler id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) deleteHandler(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.Handler().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

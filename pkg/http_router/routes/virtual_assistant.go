package routes

import (
	"errors"
	"fmt"
	"net/http"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	vaTY "github.com/mycontroller-org/server/v2/plugin/virtual_assistant/types"
)

// Registers virtual assistant api
func (h *Routes) registerVirtualAssistantRoutes() {
	h.router.HandleFunc("/api/virtualassistant", h.listVirtualAssistant).Methods(http.MethodGet)
	h.router.HandleFunc("/api/virtualassistant/{id}", h.getVirtualAssistant).Methods(http.MethodGet)
	h.router.HandleFunc("/api/virtualassistant", h.updateVirtualAssistant).Methods(http.MethodPost)
	h.router.HandleFunc("/api/virtualassistant/enable", h.enableVirtualAssistant).Methods(http.MethodPost)
	h.router.HandleFunc("/api/virtualassistant/disable", h.disableVirtualAssistant).Methods(http.MethodPost)
	h.router.HandleFunc("/api/virtualassistant", h.deleteVirtualAssistant).Methods(http.MethodDelete)
}

func (h *Routes) listVirtualAssistant(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntityVirtualAssistant, &[]vaTY.Config{})
}

func (h *Routes) getVirtualAssistant(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityVirtualAssistant, &vaTY.Config{})
}

func (h *Routes) updateVirtualAssistant(w http.ResponseWriter, r *http.Request) {
	entity := &vaTY.Config{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be an empty", http.StatusBadRequest)
		return
	}
	err = h.api.VirtualAssistant().SaveAndReload(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) deleteVirtualAssistant(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.VirtualAssistant().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

func (h *Routes) enableVirtualAssistant(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.VirtualAssistant().Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a task id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) disableVirtualAssistant(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.VirtualAssistant().Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a task id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

package routes

import (
	"errors"
	"fmt"
	"net/http"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterSourceRoutes registers source api
func (h *Routes) registerSourceRoutes() {
	h.router.HandleFunc("/api/source", h.listSources).Methods(http.MethodGet)
	h.router.HandleFunc("/api/source/{id}", h.getSource).Methods(http.MethodGet)
	h.router.HandleFunc("/api/source", h.updateSource).Methods(http.MethodPost)
	h.router.HandleFunc("/api/source", h.deleteSources).Methods(http.MethodDelete)
}

func (h *Routes) listSources(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntitySource, &[]sourceTY.Source{})
}

func (h *Routes) getSource(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntitySource, &sourceTY.Source{})
}

func (h *Routes) updateSource(w http.ResponseWriter, r *http.Request) {
	entity := &sourceTY.Source{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be empty", http.StatusBadRequest)
		return
	}
	err = h.api.Source().Save(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) deleteSources(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.Source().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

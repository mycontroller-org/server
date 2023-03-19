package routes

import (
	"errors"
	"fmt"
	"net/http"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	dataRepositoryTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterFieldRoutes registers data repository api
func (h *Routes) registerDataRepositoryRoutes() {
	h.router.HandleFunc("/api/datarepository", h.listDataRepositoryItems).Methods(http.MethodGet)
	h.router.HandleFunc("/api/datarepository/{id}", h.getDataRepositoryItem).Methods(http.MethodGet)
	h.router.HandleFunc("/api/datarepository", h.updateDataRepositoryItem).Methods(http.MethodPost)
	h.router.HandleFunc("/api/datarepository", h.deleteDataRepositoryItems).Methods(http.MethodDelete)
}

func (h *Routes) listDataRepositoryItems(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntityDataRepository, &[]dataRepositoryTY.Config{})
}

func (h *Routes) getDataRepositoryItem(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityDataRepository, &dataRepositoryTY.Config{})
}

func (h *Routes) updateDataRepositoryItem(w http.ResponseWriter, r *http.Request) {
	entity := &dataRepositoryTY.Config{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be empty", http.StatusBadRequest)
		return
	}
	err = h.api.DataRepository().Save(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) deleteDataRepositoryItems(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.DataRepository().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

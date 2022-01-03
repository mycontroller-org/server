package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	dataRepositoryAPI "github.com/mycontroller-org/server/v2/pkg/api/data_repository"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	dataRepositoryTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// RegisterFieldRoutes registers data repository api
func RegisterDataRepositoryRoutes(router *mux.Router) {
	router.HandleFunc("/api/datarepository", listDataRepositoryItems).Methods(http.MethodGet)
	router.HandleFunc("/api/datarepository/{id}", getDataRepositoryItem).Methods(http.MethodGet)
	router.HandleFunc("/api/datarepository", updateDataRepositoryItem).Methods(http.MethodPost)
	router.HandleFunc("/api/datarepository", deleteDataRepositoryItems).Methods(http.MethodDelete)
}

func listDataRepositoryItems(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, types.EntityDataRepository, &[]dataRepositoryTY.Config{})
}

func getDataRepositoryItem(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, types.EntityDataRepository, &dataRepositoryTY.Config{})
}

func updateDataRepositoryItem(w http.ResponseWriter, r *http.Request) {
	entity := &dataRepositoryTY.Config{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be empty", 400)
		return
	}
	err = dataRepositoryAPI.Save(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func deleteDataRepositoryItems(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := dataRepositoryAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	sourceAPI "github.com/mycontroller-org/server/v2/pkg/api/source"
	"github.com/mycontroller-org/server/v2/pkg/model"
	sourceML "github.com/mycontroller-org/server/v2/pkg/model/source"
	storageML "github.com/mycontroller-org/server/v2/plugin/database/storage"
)

// RegisterSourceRoutes registers source api
func RegisterSourceRoutes(router *mux.Router) {
	router.HandleFunc("/api/source", listSources).Methods(http.MethodGet)
	router.HandleFunc("/api/source/{id}", getSource).Methods(http.MethodGet)
	router.HandleFunc("/api/source", updateSource).Methods(http.MethodPost)
	router.HandleFunc("/api/source", deleteSources).Methods(http.MethodDelete)
}

func listSources(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, model.EntitySource, &[]sourceML.Source{})
}

func getSource(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, model.EntitySource, &sourceML.Source{})
}

func updateSource(w http.ResponseWriter, r *http.Request) {
	entity := &sourceML.Source{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be empty", 400)
		return
	}
	err = sourceAPI.Save(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func deleteSources(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageML.Filter, p *storageML.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := sourceAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

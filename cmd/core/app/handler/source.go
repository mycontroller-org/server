package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	sourceAPI "github.com/mycontroller-org/backend/v2/pkg/api/source"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	sourceML "github.com/mycontroller-org/backend/v2/pkg/model/source"
	storageML "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func registerSourceRoutes(router *mux.Router) {
	router.HandleFunc("/api/source", listSources).Methods(http.MethodGet)
	router.HandleFunc("/api/source/{id}", getSource).Methods(http.MethodGet)
	router.HandleFunc("/api/source", updateSource).Methods(http.MethodPost)
	router.HandleFunc("/api/source", deleteSources).Methods(http.MethodDelete)
}

func listSources(w http.ResponseWriter, r *http.Request) {
	FindMany(w, r, model.EntitySource, &[]sourceML.Source{})
}

func getSource(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, model.EntitySource, &sourceML.Source{})
}

func updateSource(w http.ResponseWriter, r *http.Request) {
	entity := &sourceML.Source{}
	err := LoadEntity(w, r, entity)
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
	UpdateData(w, r, &IDs, updateFn)
}

package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	dataRepositoryAPI "github.com/mycontroller-org/backend/v2/pkg/api/data_repository"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	dataRepositoryML "github.com/mycontroller-org/backend/v2/pkg/model/data_repository"
	stgML "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func registerDataRepositoryRoutes(router *mux.Router) {
	router.HandleFunc("/api/datarepository", listDataRepositoryItems).Methods(http.MethodGet)
	router.HandleFunc("/api/datarepository/{id}", getDataRepositoryItem).Methods(http.MethodGet)
	router.HandleFunc("/api/datarepository", updateDataRepositoryItem).Methods(http.MethodPost)
	router.HandleFunc("/api/datarepository", deleteDataRepositoryItems).Methods(http.MethodDelete)
}

func listDataRepositoryItems(w http.ResponseWriter, r *http.Request) {
	FindMany(w, r, model.EntityDataRepository, &[]dataRepositoryML.Config{})
}

func getDataRepositoryItem(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, model.EntityDataRepository, &dataRepositoryML.Config{})
}

func updateDataRepositoryItem(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]stgML.Filter) error {
		e := d.(*dataRepositoryML.Config)
		if e.ID == "" {
			return errors.New("id field should not be empty")
		}
		return nil
	}
	SaveEntity(w, r, model.EntityDataRepository, &dataRepositoryML.Config{}, bwFunc)
}

func deleteDataRepositoryItems(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := dataRepositoryAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	UpdateData(w, r, &IDs, updateFn)
}

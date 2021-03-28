package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func registerFieldRoutes(router *mux.Router) {
	router.HandleFunc("/api/field", listFields).Methods(http.MethodGet)
	router.HandleFunc("/api/field/{id}", getField).Methods(http.MethodGet)
	router.HandleFunc("/api/field", updateField).Methods(http.MethodPost)
	router.HandleFunc("/api/field", deleteFields).Methods(http.MethodDelete)
}

func listFields(w http.ResponseWriter, r *http.Request) {
	FindMany(w, r, ml.EntityField, &[]fml.Field{})
}

func getField(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, ml.EntityField, &fml.Field{})
}

func updateField(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]stgml.Filter) error {
		e := d.(*fml.Field)
		if e.ID == "" {
			return errors.New("id should not be an empty")
		}
		return nil
	}
	SaveEntity(w, r, ml.EntityField, &fml.Field{}, bwFunc)
}

func deleteFields(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := fieldAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	UpdateData(w, r, &IDs, updateFn)
}

package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/backend/v2/cmd/core/app/handler/utils"
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	fieldML "github.com/mycontroller-org/backend/v2/pkg/model/field"
	stgML "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// RegisterFieldRoutes registers field api
func RegisterFieldRoutes(router *mux.Router) {
	router.HandleFunc("/api/field", listFields).Methods(http.MethodGet)
	router.HandleFunc("/api/field/{id}", getField).Methods(http.MethodGet)
	router.HandleFunc("/api/field", updateField).Methods(http.MethodPost)
	router.HandleFunc("/api/field", deleteFields).Methods(http.MethodDelete)
}

func listFields(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, model.EntityField, &[]fieldML.Field{})
}

func getField(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, model.EntityField, &fieldML.Field{})
}

func updateField(w http.ResponseWriter, r *http.Request) {
	entity := &fieldML.Field{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be empty", 400)
		return
	}
	err = fieldAPI.Save(entity, true)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func deleteFields(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := fieldAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

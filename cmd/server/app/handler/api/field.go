package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	fieldAPI "github.com/mycontroller-org/server/v2/pkg/api/field"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterFieldRoutes registers field api
func RegisterFieldRoutes(router *mux.Router) {
	router.HandleFunc("/api/field", listFields).Methods(http.MethodGet)
	router.HandleFunc("/api/field/{id}", getField).Methods(http.MethodGet)
	router.HandleFunc("/api/field", updateField).Methods(http.MethodPost)
	router.HandleFunc("/api/field", deleteFields).Methods(http.MethodDelete)
}

func listFields(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, types.EntityField, &[]fieldTY.Field{})
}

func getField(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, types.EntityField, &fieldTY.Field{})
}

func updateField(w http.ResponseWriter, r *http.Request) {
	entity := &fieldTY.Field{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
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
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
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

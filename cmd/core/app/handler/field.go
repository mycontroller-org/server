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

func registerSensorFieldRoutes(router *mux.Router) {
	router.HandleFunc("/api/sensorfield", listSensorFields).Methods(http.MethodGet)
	router.HandleFunc("/api/sensorfield/{id}", getSensorField).Methods(http.MethodGet)
	router.HandleFunc("/api/sensorfield", updateSensorField).Methods(http.MethodPost)
	router.HandleFunc("/api/sensorfield", deleteFields).Methods(http.MethodDelete)
}

func listSensorFields(w http.ResponseWriter, r *http.Request) {
	FindMany(w, r, ml.EntitySensorField, &[]fml.Field{})
}

func getSensorField(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, ml.EntitySensorField, &fml.Field{})
}

func updateSensorField(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]stgml.Filter) error {
		e := d.(*fml.Field)
		if e.ID == "" {
			return errors.New("id should not be an empty")
		}
		return nil
	}
	SaveEntity(w, r, ml.EntitySensorField, &fml.Field{}, bwFunc)
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

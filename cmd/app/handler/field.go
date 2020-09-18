package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	fieldAPI "github.com/mycontroller-org/backend/v2/pkg/api/field"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
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
	bwFunc := func(d interface{}, f *[]pml.Filter) error {
		e := d.(*fml.Field)
		if e.ID == "" {
			return errors.New("ID should not be an empty")
		}
		return nil
	}
	SaveEntity(w, r, ml.EntitySensorField, &fml.Field{}, bwFunc)
}

func deleteFields(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []pml.Filter, p *pml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := fieldAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("Deleted: %d", count), nil
		}
		return nil, errors.New("Supply id(s)")
	}
	UpdateData(w, r, &IDs, updateFn)
}

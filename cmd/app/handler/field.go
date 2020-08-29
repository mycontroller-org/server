package handler

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
)

func registerSensorFieldRoutes(router *mux.Router) {
	router.HandleFunc("/api/sensorfield", listSensorFields).Methods(http.MethodGet)
	router.HandleFunc("/api/sensorfield/{id}", getSensorField).Methods(http.MethodGet)
	router.HandleFunc("/api/sensorfield", updateSensorField).Methods(http.MethodPost)
}

func listSensorFields(w http.ResponseWriter, r *http.Request) {
	findMany(w, r, ml.EntityField, &[]fml.Field{})
}

func getSensorField(w http.ResponseWriter, r *http.Request) {
	findOne(w, r, ml.EntityField, &fml.Field{})
}

func updateSensorField(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]pml.Filter) error {
		e := d.(*fml.Field)
		if e.ID == "" {
			return errors.New("ID should not be an empty")
		}
		return nil
	}
	saveEntity(w, r, ml.EntityField, &fml.Field{}, bwFunc)
}

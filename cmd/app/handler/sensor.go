package handler

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	sml "github.com/mycontroller-org/backend/v2/pkg/model/sensor"
)

func registerSensorRoutes(router *mux.Router) {
	router.HandleFunc("/api/sensor", listSensors).Methods(http.MethodGet)
	router.HandleFunc("/api/sensor/{id}", getSensor).Methods(http.MethodGet)
	router.HandleFunc("/api/sensor", updateSensor).Methods(http.MethodPost)
}

func listSensors(w http.ResponseWriter, r *http.Request) {
	findMany(w, r, ml.EntitySensor, &[]sml.Sensor{})
}

func getSensor(w http.ResponseWriter, r *http.Request) {
	findOne(w, r, ml.EntitySensor, &sml.Sensor{})
}

func updateSensor(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]pml.Filter) error {
		e := d.(*sml.Sensor)
		if e.ID == "" {
			return errors.New("ID field should not be empty")
		}
		return nil
	}
	saveEntity(w, r, ml.EntitySensor, &sml.Sensor{}, bwFunc)
}

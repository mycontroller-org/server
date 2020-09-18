package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	sensorAPI "github.com/mycontroller-org/backend/v2/pkg/api/sensor"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	sml "github.com/mycontroller-org/backend/v2/pkg/model/sensor"
)

func registerSensorRoutes(router *mux.Router) {
	router.HandleFunc("/api/sensor", listSensors).Methods(http.MethodGet)
	router.HandleFunc("/api/sensor/{id}", getSensor).Methods(http.MethodGet)
	router.HandleFunc("/api/sensor", updateSensor).Methods(http.MethodPost)
	router.HandleFunc("/api/sensor", deleteSensors).Methods(http.MethodDelete)
}

func listSensors(w http.ResponseWriter, r *http.Request) {
	FindMany(w, r, ml.EntitySensor, &[]sml.Sensor{})
}

func getSensor(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, ml.EntitySensor, &sml.Sensor{})
}

func updateSensor(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]pml.Filter) error {
		e := d.(*sml.Sensor)
		if e.ID == "" {
			return errors.New("ID field should not be empty")
		}
		return nil
	}
	SaveEntity(w, r, ml.EntitySensor, &sml.Sensor{}, bwFunc)
}

func deleteSensors(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []pml.Filter, p *pml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := sensorAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("Deleted: %d", count), nil
		}
		return nil, errors.New("Supply a sensor id(s)")
	}
	UpdateData(w, r, &IDs, updateFn)
}

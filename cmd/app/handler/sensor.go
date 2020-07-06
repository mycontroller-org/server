package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	ml "github.com/mycontroller-org/mycontroller-v2/pkg/model"
)

func registerSensorRoutes(router *mux.Router) {
	router.HandleFunc("/api/sensor", listSensors).Methods(http.MethodGet)
	router.HandleFunc("/api/sensor/{id}", getSensor).Methods(http.MethodGet)
	router.HandleFunc("/api/sensor", updateSensor).Methods(http.MethodPost)
}

func listSensors(w http.ResponseWriter, r *http.Request) {
	findMany(w, r, ml.EntitySensor, &[]ml.Sensor{})
}

func getSensor(w http.ResponseWriter, r *http.Request) {
	findOne(w, r, ml.EntitySensor, &ml.Sensor{})
}

func updateSensor(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]ml.Filter) error {
		e := d.(*ml.Sensor)
		if e.ID == "" {
			if e.GatewayID == "" || e.NodeID == "" || e.ShortID == "" {
				return errors.New("GatewayID, NodeId or ShortID field should not be empty")
			}
			e.ID = fmt.Sprintf("%s_%s_%s", e.GatewayID, e.NodeID, e.ShortID)
		}
		return nil
	}
	saveEntity(w, r, ml.EntitySensor, &ml.Sensor{}, bwFunc)
}

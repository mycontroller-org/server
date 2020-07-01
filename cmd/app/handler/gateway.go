package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	ml "github.com/mycontroller-org/mycontroller/pkg/model"
	ut "github.com/mycontroller-org/mycontroller/pkg/util"
)

func registerGatewayRoutes(router *mux.Router) {
	router.HandleFunc("/api/gateways", listGateways).Methods(http.MethodGet)
	router.HandleFunc("/api/gateways/{id}", getGateway).Methods(http.MethodGet)
	router.HandleFunc("/api/gateways", updateGateway).Methods(http.MethodPost)
}

func listGateways(w http.ResponseWriter, r *http.Request) {
	findMany(w, r, ml.EntityGateway, &[]ml.GatewayConfig{})
}

func getGateway(w http.ResponseWriter, r *http.Request) {
	findOne(w, r, ml.EntityGateway, &ml.GatewayConfig{})
}

func updateGateway(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]ml.Filter) error {
		e := d.(*ml.GatewayConfig)
		if e.ID == "" {
			e.ID = ut.RandID()
		}
		return nil
	}
	saveEntity(w, r, ml.EntityGateway, &ml.GatewayConfig{}, bwFunc)
}

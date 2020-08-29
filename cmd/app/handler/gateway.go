package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
	ut "github.com/mycontroller-org/backend/v2/pkg/util"
)

func registerGatewayRoutes(router *mux.Router) {
	router.HandleFunc("/api/gateway", listGateways).Methods(http.MethodGet)
	router.HandleFunc("/api/gateway/{id}", getGateway).Methods(http.MethodGet)
	router.HandleFunc("/api/gateway", updateGateway).Methods(http.MethodPost)
}

func listGateways(w http.ResponseWriter, r *http.Request) {
	findMany(w, r, ml.EntityGateway, &[]gwml.Config{})
}

func getGateway(w http.ResponseWriter, r *http.Request) {
	findOne(w, r, ml.EntityGateway, &gwml.Config{})
}

func updateGateway(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]pml.Filter) error {
		e := d.(*gwml.Config)
		if e.ID == "" {
			e.ID = ut.RandID()
		}
		return nil
	}
	saveEntity(w, r, ml.EntityGateway, &gwml.Config{}, bwFunc)
}

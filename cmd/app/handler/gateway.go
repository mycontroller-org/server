package handler

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	gwAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func registerGatewayRoutes(router *mux.Router) {
	router.HandleFunc("/api/gateway", listGateways).Methods(http.MethodGet)
	router.HandleFunc("/api/gateway/{id}", getGateway).Methods(http.MethodGet)
	router.HandleFunc("/api/gateway", updateGateway).Methods(http.MethodPost)
	router.HandleFunc("/api/gateway/enable", enableGateway).Methods(http.MethodPost)
	router.HandleFunc("/api/gateway/disable", disableGateway).Methods(http.MethodPost)
	router.HandleFunc("/api/gateway/reload", reloadGateway).Methods(http.MethodPost)
}

func listGateways(w http.ResponseWriter, r *http.Request) {
	entityFn := func(f []stgml.Filter, p *stgml.Pagination) (interface{}, error) {
		return gwAPI.List(f, p)
	}
	LoadData(w, r, entityFn)
}

func getGateway(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, ml.EntityGateway, &gwml.Config{})
}

func updateGateway(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]stgml.Filter) error {
		e := d.(*gwml.Config)
		if e.ID == "" {
			e.ID = ut.RandID()
		}
		return nil
	}
	SaveEntity(w, r, ml.EntityGateway, &gwml.Config{}, bwFunc)
}

func enableGateway(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			err := gwAPI.Enable(IDs[0])
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("Supply a gateway id")
	}
	UpdateData(w, r, &IDs, updateFn)
}

func disableGateway(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			err := gwAPI.Disable(IDs[0])
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("Supply a gateway id")
	}
	UpdateData(w, r, &IDs, updateFn)
}

func reloadGateway(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			err := gwAPI.Reload(IDs[0])
			if err != nil {
				return nil, err
			}
			return "Reloaded", nil
		}
		return nil, errors.New("Supply a gateway id")
	}
	UpdateData(w, r, &IDs, updateFn)
}

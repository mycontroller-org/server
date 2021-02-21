package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	gwAPI "github.com/mycontroller-org/backend/v2/pkg/api/gateway"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	gwml "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func registerGatewayRoutes(router *mux.Router) {
	router.HandleFunc("/api/gateway", listGateways).Methods(http.MethodGet)
	router.HandleFunc("/api/gateway/{id}", getGateway).Methods(http.MethodGet)
	router.HandleFunc("/api/gateway", updateGateway).Methods(http.MethodPost)
	router.HandleFunc("/api/gateway/enable", enableGateway).Methods(http.MethodPost)
	router.HandleFunc("/api/gateway/disable", disableGateway).Methods(http.MethodPost)
	router.HandleFunc("/api/gateway/reload", reloadGateway).Methods(http.MethodPost)
	router.HandleFunc("/api/gateway", deleteGateways).Methods(http.MethodDelete)
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
	entity := &gwml.Config{}
	err := LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "ID should not be an empty", 400)
		return
	}
	err = gwAPI.SaveAndReload(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func enableGateway(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := gwAPI.Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("Supply a gateway id")
	}
	UpdateData(w, r, &ids, updateFn)
}

func disableGateway(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := gwAPI.Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("Supply a gateway id")
	}
	UpdateData(w, r, &ids, updateFn)
}

func reloadGateway(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := gwAPI.Reload(ids)
			if err != nil {
				return nil, err
			}
			return "Reloaded", nil
		}
		return nil, errors.New("Supply a gateway id")
	}
	UpdateData(w, r, &ids, updateFn)
}

func deleteGateways(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := gwAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("Deleted: %d", count), nil
		}
		return nil, errors.New("Supply id(s)")
	}
	UpdateData(w, r, &IDs, updateFn)
}

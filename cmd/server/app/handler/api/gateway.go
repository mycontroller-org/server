package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	gwAPI "github.com/mycontroller-org/server/v2/pkg/api/gateway"
	"github.com/mycontroller-org/server/v2/pkg/model"
	gwML "github.com/mycontroller-org/server/v2/pkg/model/gateway"
	stgML "github.com/mycontroller-org/server/v2/plugin/storage"
)

// RegisterGatewayRoutes registers gateway api
func RegisterGatewayRoutes(router *mux.Router) {
	router.HandleFunc("/api/gateway", listGateways).Methods(http.MethodGet)
	router.HandleFunc("/api/gateway/{id}", getGateway).Methods(http.MethodGet)
	router.HandleFunc("/api/gateway", updateGateway).Methods(http.MethodPost)
	router.HandleFunc("/api/gateway/enable", enableGateway).Methods(http.MethodPost)
	router.HandleFunc("/api/gateway/disable", disableGateway).Methods(http.MethodPost)
	router.HandleFunc("/api/gateway/reload", reloadGateway).Methods(http.MethodPost)
	router.HandleFunc("/api/gateway", deleteGateways).Methods(http.MethodDelete)
}

func listGateways(w http.ResponseWriter, r *http.Request) {
	entityFn := func(f []stgML.Filter, p *stgML.Pagination) (interface{}, error) {
		return gwAPI.List(f, p)
	}
	handlerUtils.LoadData(w, r, entityFn)
}

func getGateway(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, model.EntityGateway, &gwML.Config{})
}

func updateGateway(w http.ResponseWriter, r *http.Request) {
	entity := &gwML.Config{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be an empty", 400)
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
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := gwAPI.Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a gateway id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func disableGateway(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := gwAPI.Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a gateway id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func reloadGateway(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := gwAPI.Reload(ids)
			if err != nil {
				return nil, err
			}
			return "Reloaded", nil
		}
		return nil, errors.New("supply a gateway id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func deleteGateways(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := gwAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

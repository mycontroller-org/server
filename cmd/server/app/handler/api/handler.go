package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	handlerAPI "github.com/mycontroller-org/server/v2/pkg/api/handler"
	"github.com/mycontroller-org/server/v2/pkg/model"
	handlerType "github.com/mycontroller-org/server/v2/plugin/handler/type"
	stgML "github.com/mycontroller-org/server/v2/plugin/database/storage"
)

// RegisterHandlerRoutes registers handler api
func RegisterHandlerRoutes(router *mux.Router) {
	router.HandleFunc("/api/handler", listHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/handler/{id}", getHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/handler", updateHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/handler/enable", enableHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/handler/disable", disableHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/handler/reload", reloadHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/handler", deleteHandler).Methods(http.MethodDelete)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, model.EntityHandler, &[]handlerType.Config{})
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, model.EntityHandler, &handlerType.Config{})
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	entity := &handlerType.Config{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be an empty", 400)
		return
	}
	err = handlerAPI.SaveAndReload(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func enableHandler(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := handlerAPI.Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a handler id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func disableHandler(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := handlerAPI.Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a handler id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func reloadHandler(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := handlerAPI.Reload(ids)
			if err != nil {
				return nil, err
			}
			return "Reloaded", nil
		}
		return nil, errors.New("supply a handler id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := handlerAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

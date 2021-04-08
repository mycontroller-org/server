package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerAPI "github.com/mycontroller-org/backend/v2/pkg/api/handler"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func registerNotifyHandlerRoutes(router *mux.Router) {
	router.HandleFunc("/api/handler", listNotifyHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/handler/{id}", getNotifyHandler).Methods(http.MethodGet)
	router.HandleFunc("/api/handler", updateNotifyHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/handler/enable", enableNotifyHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/handler/disable", disableNotifyHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/handler/reload", reloadNotifyHandler).Methods(http.MethodPost)
	router.HandleFunc("/api/handler", deleteNotifyHandler).Methods(http.MethodDelete)
}

func listNotifyHandler(w http.ResponseWriter, r *http.Request) {
	FindMany(w, r, ml.EntityHandler, &[]handlerML.Config{})
}

func getNotifyHandler(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, ml.EntityHandler, &handlerML.Config{})
}

func updateNotifyHandler(w http.ResponseWriter, r *http.Request) {
	entity := &handlerML.Config{}
	err := LoadEntity(w, r, entity)
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

func enableNotifyHandler(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := handlerAPI.Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a handler id")
	}
	UpdateData(w, r, &ids, updateFn)
}

func disableNotifyHandler(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := handlerAPI.Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a handler id")
	}
	UpdateData(w, r, &ids, updateFn)
}

func reloadNotifyHandler(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := handlerAPI.Reload(ids)
			if err != nil {
				return nil, err
			}
			return "Reloaded", nil
		}
		return nil, errors.New("supply a handler id")
	}
	UpdateData(w, r, &ids, updateFn)
}

func deleteNotifyHandler(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := handlerAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	UpdateData(w, r, &IDs, updateFn)
}

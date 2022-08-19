package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	vaAPI "github.com/mycontroller-org/server/v2/pkg/api/virtual_assistant"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	vaTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_assistant"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// Registers virtual assistant api
func RegisterVirtualAssistantRoutes(router *mux.Router) {
	router.HandleFunc("/api/virtualassistant", listVirtualAssistant).Methods(http.MethodGet)
	router.HandleFunc("/api/virtualassistant/{id}", getVirtualAssistant).Methods(http.MethodGet)
	router.HandleFunc("/api/virtualassistant", updateVirtualAssistant).Methods(http.MethodPost)
	router.HandleFunc("/api/virtualassistant/enable", enableVirtualAssistant).Methods(http.MethodPost)
	router.HandleFunc("/api/virtualassistant/disable", disableVirtualAssistant).Methods(http.MethodPost)
	router.HandleFunc("/api/virtualassistant", deleteVirtualAssistant).Methods(http.MethodDelete)
}

func listVirtualAssistant(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, types.EntityVirtualAssistant, &[]vaTY.Config{})
}

func getVirtualAssistant(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, types.EntityVirtualAssistant, &vaTY.Config{})
}

func updateVirtualAssistant(w http.ResponseWriter, r *http.Request) {
	entity := &vaTY.Config{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be an empty", 400)
		return
	}
	err = vaAPI.SaveAndReload(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func deleteVirtualAssistant(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := vaAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

func enableVirtualAssistant(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := vaAPI.Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a task id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func disableVirtualAssistant(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := vaAPI.Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a task id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

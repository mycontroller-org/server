package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	middleware "github.com/mycontroller-org/server/v2/cmd/server/app/handler/middleware"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	svcTokenAPI "github.com/mycontroller-org/server/v2/pkg/api/service_token"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// registers service token routes
func RegisterServiceTokenRoutes(router *mux.Router) {
	router.HandleFunc("/api/servicetoken", listServiceToken).Methods(http.MethodGet)
	router.HandleFunc("/api/servicetoken/{id}", getServiceToken).Methods(http.MethodGet)
	router.HandleFunc("/api/servicetoken/create", createServiceToken).Methods(http.MethodPost)
	router.HandleFunc("/api/servicetoken/update", updateServiceToken).Methods(http.MethodPost)
	router.HandleFunc("/api/servicetoken", deleteServiceToken).Methods(http.MethodDelete)
}

func listServiceToken(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, types.EntityServiceToken, &[]svcTokenTY.ServiceToken{})
}

func getServiceToken(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, types.EntityServiceToken, &svcTokenTY.ServiceToken{})
}

func updateServiceToken(w http.ResponseWriter, r *http.Request) {
	entity := &svcTokenTY.ServiceToken{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// update userId
	entity.UserID = middleware.GetUserID(r)

	err = svcTokenAPI.Save(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func createServiceToken(w http.ResponseWriter, r *http.Request) {
	entity := &svcTokenTY.ServiceToken{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// update userId
	entity.UserID = middleware.GetUserID(r)

	generatedToken, err := svcTokenAPI.Create(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return generated token
	handlerUtils.PostSuccessResponse(w, generatedToken)
}

func deleteServiceToken(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := svcTokenAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

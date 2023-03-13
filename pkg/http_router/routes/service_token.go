package routes

import (
	"errors"
	"fmt"
	"net/http"

	middleware "github.com/mycontroller-org/server/v2/pkg/http_router/middleware"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	svcTokenTY "github.com/mycontroller-org/server/v2/pkg/types/service_token"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// registers service token routes
func (h *Routes) registerServiceTokenRoutes() {
	h.router.HandleFunc("/api/servicetoken", h.listServiceToken).Methods(http.MethodGet)
	h.router.HandleFunc("/api/servicetoken/{id}", h.getServiceToken).Methods(http.MethodGet)
	h.router.HandleFunc("/api/servicetoken/create", h.createServiceToken).Methods(http.MethodPost)
	h.router.HandleFunc("/api/servicetoken/update", h.updateServiceToken).Methods(http.MethodPost)
	h.router.HandleFunc("/api/servicetoken", h.deleteServiceToken).Methods(http.MethodDelete)
}

func (h *Routes) listServiceToken(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntityServiceToken, &[]svcTokenTY.ServiceToken{})
}

func (h *Routes) getServiceToken(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityServiceToken, &svcTokenTY.ServiceToken{})
}

func (h *Routes) updateServiceToken(w http.ResponseWriter, r *http.Request) {
	entity := &svcTokenTY.ServiceToken{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// update userId
	entity.UserID = middleware.GetUserID(r)

	err = h.api.ServiceToken().Save(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) createServiceToken(w http.ResponseWriter, r *http.Request) {
	entity := &svcTokenTY.ServiceToken{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// update userId
	entity.UserID = middleware.GetUserID(r)

	generatedToken, err := h.api.ServiceToken().Create(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return generated token
	handlerUtils.PostSuccessResponse(w, generatedToken)
}

func (h *Routes) deleteServiceToken(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.ServiceToken().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

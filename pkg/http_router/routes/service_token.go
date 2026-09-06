package routes

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
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

// Service tokens are personal credentials: a token acts as its owner. They are
// therefore always scoped to the caller, whatever the caller's policies say -
// otherwise one principal could read, widen (drop the action/resource limits, set
// neverExpire) or delete another principal's credentials.
func (h *Routes) ownedByCaller(r *http.Request) []storageTY.Filter {
	return []storageTY.Filter{{Key: types.KeyUserID, Value: middleware.GetUserID(r)}}
}

func (h *Routes) listServiceToken(w http.ResponseWriter, r *http.Request) {
	entityFn := func(f []storageTY.Filter, p *storageTY.Pagination) (interface{}, error) {
		return h.api.ServiceToken().List(append(f, h.ownedByCaller(r)...), p)
	}
	handlerUtils.LoadData(w, r, entityFn)
}

func (h *Routes) getServiceToken(w http.ResponseWriter, r *http.Request) {
	token, err := h.callerToken(r, mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	handlerUtils.PostSuccessResponse(w, token)
}

// callerToken loads a token and verifies the caller owns it.
func (h *Routes) callerToken(r *http.Request, id string) (*svcTokenTY.ServiceToken, error) {
	if id == "" {
		return nil, errors.New("id should not be an empty")
	}
	token, err := h.api.ServiceToken().GetByID(id)
	if err != nil {
		return nil, err
	}
	if token.UserID != middleware.GetUserID(r) {
		// do not disclose that the id exists
		return nil, errors.New("service token not found")
	}
	return &token, nil
}

func (h *Routes) updateServiceToken(w http.ResponseWriter, r *http.Request) {
	entity := &svcTokenTY.ServiceToken{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := h.callerToken(r, entity.ID); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
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
		if len(IDs) == 0 {
			return nil, errors.New("supply id(s)")
		}
		// only the owner's tokens may be deleted
		for _, id := range IDs {
			if _, err := h.callerToken(r, id); err != nil {
				return nil, err
			}
		}
		count, err := h.api.ServiceToken().Delete(IDs)
		if err != nil {
			return nil, err
		}
		return fmt.Sprintf("deleted: %d", count), nil
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

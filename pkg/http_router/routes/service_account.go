package routes

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	middleware "github.com/mycontroller-org/server/v2/pkg/http_router/middleware"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	svcAccountTY "github.com/mycontroller-org/server/v2/pkg/types/service_account"
	userTY "github.com/mycontroller-org/server/v2/pkg/types/user"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// registers service account routes
func (h *Routes) registerServiceAccountRoutes() {
	h.router.HandleFunc("/api/serviceaccount", h.listServiceAccount).Methods(http.MethodGet)
	h.router.HandleFunc("/api/serviceaccount/{id}", h.getServiceAccount).Methods(http.MethodGet)
	h.router.HandleFunc("/api/serviceaccount/create", h.createServiceAccount).Methods(http.MethodPost)
	h.router.HandleFunc("/api/serviceaccount/update", h.updateServiceAccount).Methods(http.MethodPost)
	h.router.HandleFunc("/api/serviceaccount", h.deleteServiceAccount).Methods(http.MethodDelete)
}

// ownedByCaller limits list results to the logged-in user. Callers who can
// manage users (kind-wide user update) see every service account.
func (h *Routes) ownedByCaller(r *http.Request) []storageTY.Filter {
	return []storageTY.Filter{{Key: types.KeyUserID, Value: middleware.GetUserID(r)}}
}

func (h *Routes) canManageOtherUsers(r *http.Request) bool {
	subject, err := middleware.SubjectFromRequest(r)
	if err != nil {
		return false
	}
	return h.api.Policy().AllowedKindWide(subject, policyTY.ActionUpdate, policyTY.ResourceUser) == nil
}

func (h *Routes) lookupUser(ref string) (*userTY.User, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, errors.New("user not found")
	}
	if user, err := h.api.User().GetByID(ref); err == nil && user.ID != "" {
		return &user, nil
	}
	if user, err := h.api.User().GetByUsername(ref); err == nil && user.ID != "" {
		return &user, nil
	}
	return nil, errors.New("user not found")
}

// resolveOwner returns the user id the service account should belong to.
// Empty userId/username means the caller. Creating for someone else requires
// kind-wide user update (admin / user manager).
func (h *Routes) resolveOwner(r *http.Request, userID, username string) (*userTY.User, error) {
	callerID := middleware.GetUserID(r)
	ref := strings.TrimSpace(userID)
	if ref == "" {
		ref = strings.TrimSpace(username)
	}
	if ref == "" || ref == callerID {
		user, err := h.api.User().GetByID(callerID)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}
	user, err := h.lookupUser(ref)
	if err != nil {
		return nil, err
	}
	if user.ID != callerID && !h.canManageOtherUsers(r) {
		return nil, errors.New("not allowed to create a service account for another user")
	}
	return user, nil
}

func (h *Routes) listServiceAccount(w http.ResponseWriter, r *http.Request) {
	entityFn := func(f []storageTY.Filter, p *storageTY.Pagination) (interface{}, error) {
		filters := f
		if !h.canManageOtherUsers(r) {
			filters = append(filters, h.ownedByCaller(r)...)
		}
		return h.api.ServiceAccount().List(filters, p)
	}
	handlerUtils.LoadData(w, r, entityFn)
}

func (h *Routes) getServiceAccount(w http.ResponseWriter, r *http.Request) {
	token, err := h.loadServiceAccount(r, mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	handlerUtils.PostSuccessResponse(w, token)
}

// loadServiceAccount loads an account if the caller owns it or can manage users.
func (h *Routes) loadServiceAccount(r *http.Request, id string) (*svcAccountTY.ServiceAccount, error) {
	if id == "" {
		return nil, errors.New("id should not be an empty")
	}
	token, err := h.api.ServiceAccount().GetByID(id)
	if err != nil {
		return nil, err
	}
	if token.UserID != middleware.GetUserID(r) && !h.canManageOtherUsers(r) {
		return nil, errors.New("service account not found")
	}
	return &token, nil
}

func (h *Routes) updateServiceAccount(w http.ResponseWriter, r *http.Request) {
	entity := &svcAccountTY.ServiceAccount{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	existing, err := h.loadServiceAccount(r, entity.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// owner is immutable
	entity.UserID = existing.UserID
	entity.Username = existing.Username
	if user, err := h.api.User().GetByID(existing.UserID); err == nil {
		entity.Username = user.Username
	}

	err = h.api.ServiceAccount().Save(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) createServiceAccount(w http.ResponseWriter, r *http.Request) {
	entity := &svcAccountTY.ServiceAccount{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	owner, err := h.resolveOwner(r, entity.UserID, entity.Username)
	if err != nil {
		status := http.StatusBadRequest
		if err.Error() == "not allowed to create a service account for another user" {
			status = http.StatusForbidden
		}
		http.Error(w, err.Error(), status)
		return
	}
	entity.UserID = owner.ID
	entity.Username = owner.Username

	generatedToken, err := h.api.ServiceAccount().Create(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	handlerUtils.PostSuccessResponse(w, generatedToken)
}

func (h *Routes) deleteServiceAccount(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) == 0 {
			return nil, errors.New("supply id(s)")
		}
		for _, id := range IDs {
			if _, err := h.loadServiceAccount(r, id); err != nil {
				return nil, err
			}
		}
		count, err := h.api.ServiceAccount().Delete(IDs)
		if err != nil {
			return nil, err
		}
		return fmt.Sprintf("deleted: %d", count), nil
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

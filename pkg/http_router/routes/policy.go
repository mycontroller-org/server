package routes

import (
	"errors"
	"net/http"

	policyAPI "github.com/mycontroller-org/server/v2/pkg/api/policy"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	policyTY "github.com/mycontroller-org/server/v2/pkg/types/policy"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

func (h *Routes) registerPolicyRoutes() {
	h.router.HandleFunc("/api/policy", h.listPolicies).Methods(http.MethodGet)
	h.router.HandleFunc("/api/policy/{id}", h.getPolicy).Methods(http.MethodGet)
	h.router.HandleFunc("/api/policy", h.updatePolicy).Methods(http.MethodPost)
	h.router.HandleFunc("/api/policy", h.deletePolicies).Methods(http.MethodDelete)
}

func (h *Routes) listPolicies(w http.ResponseWriter, r *http.Request) {
	entityFn := func(f []storageTY.Filter, p *storageTY.Pagination) (interface{}, error) {
		return h.api.Policy().List(f, p)
	}
	handlerUtils.LoadData(w, r, entityFn)
}

func (h *Routes) getPolicy(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityPolicy, &policyTY.Policy{})
}

func (h *Routes) updatePolicy(w http.ResponseWriter, r *http.Request) {
	entity := &policyTY.Policy{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if entity.ID == "" {
		http.Error(w, "id should not be an empty", http.StatusBadRequest)
		return
	}
	err = h.api.Policy().Save(entity)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, policyAPI.ErrSystemPolicyImmutable) || errors.Is(err, policyAPI.ErrSystemFlagNotAllowed) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
}

func (h *Routes) deletePolicies(w http.ResponseWriter, r *http.Request) {
	IDs := make([]string, 0)
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.Policy().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return count, nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

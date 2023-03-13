package routes

import (
	"errors"
	"fmt"
	"net/http"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterFieldRoutes registers field api
func (h *Routes) registerFieldRoutes() {
	h.router.HandleFunc("/api/field", h.listFields).Methods(http.MethodGet)
	h.router.HandleFunc("/api/field/{id}", h.getField).Methods(http.MethodGet)
	h.router.HandleFunc("/api/field", h.updateField).Methods(http.MethodPost)
	h.router.HandleFunc("/api/field", h.deleteFields).Methods(http.MethodDelete)
}

func (h *Routes) listFields(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntityField, &[]fieldTY.Field{})
}

func (h *Routes) getField(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityField, &fieldTY.Field{})
}

func (h *Routes) updateField(w http.ResponseWriter, r *http.Request) {
	entity := &fieldTY.Field{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.api.Field().Save(entity, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) deleteFields(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.Field().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

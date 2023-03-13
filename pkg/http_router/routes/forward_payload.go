package routes

import (
	"errors"
	"fmt"
	"net/http"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	fwdPayloadTY "github.com/mycontroller-org/server/v2/pkg/types/forward_payload"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterForwardPayloadRoutes registers forward payload api
func (h *Routes) registerForwardPayloadRoutes() {
	h.router.HandleFunc("/api/forwardpayload", h.listForwardPayload).Methods(http.MethodGet)
	h.router.HandleFunc("/api/forwardpayload/{id}", h.getForwardPayload).Methods(http.MethodGet)
	h.router.HandleFunc("/api/forwardpayload", h.updateForwardPayload).Methods(http.MethodPost)
	h.router.HandleFunc("/api/forwardpayload", h.deleteForwardPayload).Methods(http.MethodDelete)
	h.router.HandleFunc("/api/forwardpayload/enable", h.enableForwardPayload).Methods(http.MethodPost)
	h.router.HandleFunc("/api/forwardpayload/disable", h.disableForwardPayload).Methods(http.MethodPost)
}

func (h *Routes) listForwardPayload(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntityForwardPayload, &[]fwdPayloadTY.Config{})
}

func (h *Routes) getForwardPayload(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityForwardPayload, &fwdPayloadTY.Config{})
}

func (h *Routes) updateForwardPayload(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]storageTY.Filter) error {
		e := d.(*fwdPayloadTY.Config)
		if e.ID == "" {
			return errors.New("id should not be an empty")
		}
		return nil
	}
	handlerUtils.SaveEntity(h.storage, w, r, types.EntityForwardPayload, &fwdPayloadTY.Config{}, bwFunc)
}

func (h *Routes) deleteForwardPayload(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			count, err := h.api.ForwardPayload().Delete(ids)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) enableForwardPayload(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.ForwardPayload().Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a mapping id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) disableForwardPayload(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.ForwardPayload().Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a mapping id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

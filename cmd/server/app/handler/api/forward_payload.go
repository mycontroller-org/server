package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	fwdpayloadAPI "github.com/mycontroller-org/server/v2/pkg/api/forward_payload"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	fwdPayloadTY "github.com/mycontroller-org/server/v2/pkg/types/forward_payload"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/type"
)

// RegisterForwardPayloadRoutes registers forward payload api
func RegisterForwardPayloadRoutes(router *mux.Router) {
	router.HandleFunc("/api/forwardpayload", listForwardPayload).Methods(http.MethodGet)
	router.HandleFunc("/api/forwardpayload/{id}", getForwardPayload).Methods(http.MethodGet)
	router.HandleFunc("/api/forwardpayload", updateForwardPayload).Methods(http.MethodPost)
	router.HandleFunc("/api/forwardpayload", deleteForwardPayload).Methods(http.MethodDelete)
	router.HandleFunc("/api/forwardpayload/enable", enableForwardPayload).Methods(http.MethodPost)
	router.HandleFunc("/api/forwardpayload/disable", disableForwardPayload).Methods(http.MethodPost)
}

func listForwardPayload(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, types.EntityForwardPayload, &[]fwdPayloadTY.Config{})
}

func getForwardPayload(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, types.EntityForwardPayload, &fwdPayloadTY.Config{})
}

func updateForwardPayload(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]storageTY.Filter) error {
		e := d.(*fwdPayloadTY.Config)
		if e.ID == "" {
			return errors.New("id should not be an empty")
		}
		return nil
	}
	handlerUtils.SaveEntity(w, r, types.EntityForwardPayload, &fwdPayloadTY.Config{}, bwFunc)
}

func deleteForwardPayload(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			count, err := fwdpayloadAPI.Delete(ids)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func enableForwardPayload(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := fwdpayloadAPI.Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a mapping id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func disableForwardPayload(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := fwdpayloadAPI.Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a mapping id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

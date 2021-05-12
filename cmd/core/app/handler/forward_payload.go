package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	fpAPI "github.com/mycontroller-org/backend/v2/pkg/api/forward_payload"
	"github.com/mycontroller-org/backend/v2/pkg/model"
	fwdPayloadML "github.com/mycontroller-org/backend/v2/pkg/model/forward_payload"
	stgML "github.com/mycontroller-org/backend/v2/plugin/storage"
)

func registerForwardPayloadRoutes(router *mux.Router) {
	router.HandleFunc("/api/forwardpayload", listForwardPayload).Methods(http.MethodGet)
	router.HandleFunc("/api/forwardpayload/{id}", getForwardPayload).Methods(http.MethodGet)
	router.HandleFunc("/api/forwardpayload", updateForwardPayload).Methods(http.MethodPost)
	router.HandleFunc("/api/forwardpayload", deleteForwardPayload).Methods(http.MethodDelete)
	router.HandleFunc("/api/forwardpayload/enable", enableForwardPayload).Methods(http.MethodPost)
	router.HandleFunc("/api/forwardpayload/disable", disableForwardPayload).Methods(http.MethodPost)
}

func listForwardPayload(w http.ResponseWriter, r *http.Request) {
	FindMany(w, r, model.EntityForwardPayload, &[]fwdPayloadML.Config{})
}

func getForwardPayload(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, model.EntityForwardPayload, &fwdPayloadML.Config{})
}

func updateForwardPayload(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]stgML.Filter) error {
		e := d.(*fwdPayloadML.Config)
		if e.ID == "" {
			return errors.New("id should not be an empty")
		}
		return nil
	}
	SaveEntity(w, r, model.EntityForwardPayload, &fwdPayloadML.Config{}, bwFunc)
}

func deleteForwardPayload(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			count, err := fpAPI.Delete(ids)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	UpdateData(w, r, &ids, updateFn)
}

func enableForwardPayload(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := fpAPI.Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a mapping id")
	}
	UpdateData(w, r, &ids, updateFn)
}

func disableForwardPayload(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgML.Filter, p *stgML.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := fpAPI.Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a mapping id")
	}
	UpdateData(w, r, &ids, updateFn)
}

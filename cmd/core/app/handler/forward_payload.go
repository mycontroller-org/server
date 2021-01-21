package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	fpAPI "github.com/mycontroller-org/backend/v2/pkg/api/forward_payload"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	fpml "github.com/mycontroller-org/backend/v2/pkg/model/forward_payload"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
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
	FindMany(w, r, ml.EntityForwardPayload, &[]fpml.Mapping{})
}

func getForwardPayload(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, ml.EntityForwardPayload, &fpml.Mapping{})
}

func updateForwardPayload(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]stgml.Filter) error {
		e := d.(*fpml.Mapping)
		if e.ID == "" {
			return errors.New("ID should not be an empty")
		}
		return nil
	}
	SaveEntity(w, r, ml.EntityForwardPayload, &fpml.Mapping{}, bwFunc)
}

func deleteForwardPayload(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			count, err := fpAPI.Delete(ids)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("Deleted: %d", count), nil
		}
		return nil, errors.New("Supply id(s)")
	}
	UpdateData(w, r, &ids, updateFn)
}

func enableForwardPayload(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := fpAPI.Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("Supply a mapping id")
	}
	UpdateData(w, r, &ids, updateFn)
}

func disableForwardPayload(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := fpAPI.Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("Supply a mapping id")
	}
	UpdateData(w, r, &ids, updateFn)
}

package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
)

func registerNodeRoutes(router *mux.Router) {
	router.HandleFunc("/api/node", listnodes).Methods(http.MethodGet)
	router.HandleFunc("/api/node/{id}", getnode).Methods(http.MethodGet)
	router.HandleFunc("/api/node", updatenode).Methods(http.MethodPost)
	router.HandleFunc("/api/node", deleteNodes).Methods(http.MethodDelete)
}

func listnodes(w http.ResponseWriter, r *http.Request) {
	FindMany(w, r, ml.EntityNode, &[]nml.Node{})
}

func getnode(w http.ResponseWriter, r *http.Request) {
	FindOne(w, r, ml.EntityNode, &nml.Node{})
}

func updatenode(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]pml.Filter) error {
		e := d.(*nml.Node)
		if e.ID == "" {
			return errors.New("ID field should not be empty")
		}
		return nil
	}
	SaveEntity(w, r, ml.EntityNode, &nml.Node{}, bwFunc)
}

func deleteNodes(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []pml.Filter, p *pml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := nodeAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("Deleted: %d", count), nil
		}
		return nil, errors.New("Supply id(s)")
	}
	UpdateData(w, r, &IDs, updateFn)
}

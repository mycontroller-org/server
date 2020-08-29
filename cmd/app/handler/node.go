package handler

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	pml "github.com/mycontroller-org/backend/v2/pkg/model/pagination"
)

func registerNodeRoutes(router *mux.Router) {
	router.HandleFunc("/api/node", listnodes).Methods(http.MethodGet)
	router.HandleFunc("/api/node/{id}", getnode).Methods(http.MethodGet)
	router.HandleFunc("/api/node", updatenode).Methods(http.MethodPost)
}

func listnodes(w http.ResponseWriter, r *http.Request) {
	findMany(w, r, ml.EntityNode, &[]nml.Node{})
}

func getnode(w http.ResponseWriter, r *http.Request) {
	findOne(w, r, ml.EntityNode, &nml.Node{})
}

func updatenode(w http.ResponseWriter, r *http.Request) {
	bwFunc := func(d interface{}, f *[]pml.Filter) error {
		e := d.(*nml.Node)
		if e.ID == "" {
			return errors.New("ID field should not be empty")
		}
		return nil
	}
	saveEntity(w, r, ml.EntityNode, &nml.Node{}, bwFunc)
}

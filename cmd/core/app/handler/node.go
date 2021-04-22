package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	nodeAPI "github.com/mycontroller-org/backend/v2/pkg/api/node"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	nml "github.com/mycontroller-org/backend/v2/pkg/model/node"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
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
	entity := &nml.Node{}
	err := LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be empty", 400)
		return
	}
	err = nodeAPI.Save(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func deleteNodes(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []stgml.Filter, p *stgml.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := nodeAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	UpdateData(w, r, &IDs, updateFn)
}

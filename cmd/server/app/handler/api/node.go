package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	nodeAPI "github.com/mycontroller-org/server/v2/pkg/api/node"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterNodeRoutes registers node api
func RegisterNodeRoutes(router *mux.Router) {
	router.HandleFunc("/api/node", listnodes).Methods(http.MethodGet)
	router.HandleFunc("/api/node/{id}", getnode).Methods(http.MethodGet)
	router.HandleFunc("/api/node", updatenode).Methods(http.MethodPost)
	router.HandleFunc("/api/node", deleteNodes).Methods(http.MethodDelete)
}

func listnodes(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(w, r, types.EntityNode, &[]nodeTY.Node{})
}

func getnode(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(w, r, types.EntityNode, &nodeTY.Node{})
}

func updatenode(w http.ResponseWriter, r *http.Request) {
	entity := &nodeTY.Node{}
	err := handlerUtils.LoadEntity(w, r, entity)
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
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := nodeAPI.Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

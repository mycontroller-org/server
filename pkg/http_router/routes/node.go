package routes

import (
	"errors"
	"fmt"
	"net/http"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
)

// RegisterNodeRoutes registers node api
func (h *Routes) registerNodeRoutes() {
	h.router.HandleFunc("/api/node", h.listNodes).Methods(http.MethodGet)
	h.router.HandleFunc("/api/node/{id}", h.getNode).Methods(http.MethodGet)
	h.router.HandleFunc("/api/node", h.updateNode).Methods(http.MethodPost)
	h.router.HandleFunc("/api/node", h.deleteNodes).Methods(http.MethodDelete)
}

func (h *Routes) listNodes(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindMany(h.storage, w, r, types.EntityNode, &[]nodeTY.Node{})
}

func (h *Routes) getNode(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityNode, &nodeTY.Node{})
}

func (h *Routes) updateNode(w http.ResponseWriter, r *http.Request) {
	entity := &nodeTY.Node{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be empty", http.StatusBadRequest)
		return
	}
	err = h.api.Node().Save(entity, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) deleteNodes(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.Node().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

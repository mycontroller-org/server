package routes

import (
	"errors"
	"fmt"
	"net/http"

	types "github.com/mycontroller-org/server/v2/pkg/types"
	msgTY "github.com/mycontroller-org/server/v2/pkg/types/message"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gwTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
)

// RegisterGatewayRoutes registers gateway api
func (h *Routes) registerGatewayRoutes() {
	h.router.HandleFunc("/api/gateway", h.listGateways).Methods(http.MethodGet)
	h.router.HandleFunc("/api/gateway/{id}", h.getGateway).Methods(http.MethodGet)
	h.router.HandleFunc("/api/gateway", h.updateGateway).Methods(http.MethodPost)
	h.router.HandleFunc("/api/gateway/enable", h.enableGateway).Methods(http.MethodPost)
	h.router.HandleFunc("/api/gateway/disable", h.disableGateway).Methods(http.MethodPost)
	h.router.HandleFunc("/api/gateway/reload", h.reloadGateway).Methods(http.MethodPost)
	h.router.HandleFunc("/api/gateway", h.deleteGateways).Methods(http.MethodDelete)
	h.router.HandleFunc("/api/gateway-sleeping-queue", h.getSleepingQueue).Methods(http.MethodGet)
	h.router.HandleFunc("/api/gateway-sleeping-queue/clear", h.clearSleepingQueue).Methods(http.MethodGet)
}

func (h *Routes) listGateways(w http.ResponseWriter, r *http.Request) {
	entityFn := func(f []storageTY.Filter, p *storageTY.Pagination) (interface{}, error) {
		return h.api.Gateway().List(f, p)
	}
	handlerUtils.LoadData(w, r, entityFn)
}

func (h *Routes) getGateway(w http.ResponseWriter, r *http.Request) {
	handlerUtils.FindOne(h.storage, w, r, types.EntityGateway, &gwTY.Config{})
}

func (h *Routes) updateGateway(w http.ResponseWriter, r *http.Request) {
	entity := &gwTY.Config{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be an empty", 400)
		return
	}
	err = h.api.Gateway().SaveAndReload(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func (h *Routes) enableGateway(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.Gateway().Enable(ids)
			if err != nil {
				return nil, err
			}
			return "Enabled", nil
		}
		return nil, errors.New("supply a gateway id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) disableGateway(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.Gateway().Disable(ids)
			if err != nil {
				return nil, err
			}
			return "Disabled", nil
		}
		return nil, errors.New("supply a gateway id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) reloadGateway(w http.ResponseWriter, r *http.Request) {
	ids := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(ids) > 0 {
			err := h.api.Gateway().Reload(ids)
			if err != nil {
				return nil, err
			}
			return "Reloaded", nil
		}
		return nil, errors.New("supply a gateway id")
	}
	handlerUtils.UpdateData(w, r, &ids, updateFn)
}

func (h *Routes) deleteGateways(w http.ResponseWriter, r *http.Request) {
	IDs := []string{}
	updateFn := func(f []storageTY.Filter, p *storageTY.Pagination, d []byte) (interface{}, error) {
		if len(IDs) > 0 {
			count, err := h.api.Gateway().Delete(IDs)
			if err != nil {
				return nil, err
			}
			return fmt.Sprintf("deleted: %d", count), nil
		}
		return nil, errors.New("supply id(s)")
	}
	handlerUtils.UpdateData(w, r, &IDs, updateFn)
}

func (h *Routes) getSleepingQueue(w http.ResponseWriter, r *http.Request) {
	params, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gatewayID := handlerUtils.GetParameter("gatewayId", params)
	nodeID := handlerUtils.GetParameter("nodeId", params)

	if gatewayID == "" {
		http.Error(w, "gateway id can not be empty", http.StatusBadRequest)
		return
	}
	if nodeID != "" {
		messages, err := h.api.Gateway().GetNodeSleepingQueue(gatewayID, nodeID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if messages == nil {
			messages = make([]msgTY.Message, 0)
		}
		handlerUtils.PostSuccessResponse(w, messages)
		return
	} else {
		messages, err := h.api.Gateway().GetGatewaySleepingQueue(gatewayID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if messages == nil {
			messages = make(map[string][]msgTY.Message)
		}
		handlerUtils.PostSuccessResponse(w, messages)
		return
	}
}

func (h *Routes) clearSleepingQueue(w http.ResponseWriter, r *http.Request) {
	params, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	gatewayID := handlerUtils.GetParameter("gatewayId", params)
	nodeID := handlerUtils.GetParameter("nodeId", params)

	if gatewayID == "" {
		http.Error(w, "gateway id can not be empty", http.StatusBadRequest)
		return
	}
	err = h.api.Gateway().ClearSleepingQueue(gatewayID, nodeID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

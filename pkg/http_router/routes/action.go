package routes

import (
	"net/http"

	webHandlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
)

const (
	keyResource = "resource"
	keyPayload  = "payload"
	keyKeyPath  = "keyPath"
	keyAction   = "action"
	keyID       = "id"
)

// RegisterActionRoutes registers action api
func (h *Routes) registerActionRoutes() {
	h.router.HandleFunc("/api/action", h.executeGetAction).Methods(http.MethodGet)
	h.router.HandleFunc("/api/action", h.executePostAction).Methods(http.MethodPost)
	h.router.HandleFunc("/api/action/node", h.executeNodeAction).Methods(http.MethodGet)
	h.router.HandleFunc("/api/action/gateway", h.executeGatewayAction).Methods(http.MethodGet)
}

func (h *Routes) executeNodeAction(w http.ResponseWriter, r *http.Request) {
	query, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	actionArr := query[keyAction]
	idsArr := query[keyID]

	if len(actionArr) == 0 || len(idsArr) == 0 {
		http.Error(w, "required field(s) missing", http.StatusBadRequest)
		return
	}

	err = h.action.ExecuteNodeAction(actionArr[0], idsArr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) executeGatewayAction(w http.ResponseWriter, r *http.Request) {
	query, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	actionArr := query[keyAction]
	idsArr := query[keyID]

	if len(actionArr) == 0 || len(idsArr) == 0 {
		http.Error(w, "required field(s) missing", http.StatusBadRequest)
		return
	}

	err = h.action.ExecuteGatewayAction(actionArr[0], idsArr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) executeGetAction(w http.ResponseWriter, r *http.Request) {
	query, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resourceArr := query[keyResource]
	keyPathArr := query[keyKeyPath]
	payloadArr := query[keyPayload]

	if len(resourceArr) == 0 || len(payloadArr) == 0 {
		http.Error(w, "required field(s) missing", http.StatusInternalServerError)
		return
	}

	resourceData := &handlerTY.ResourceData{
		QuickID: resourceArr[0],
		Payload: payloadArr[0],
	}
	if len(keyPathArr) > 0 {
		resourceData.KeyPath = keyPathArr[0]
	}
	err = h.action.ExecuteActionOnResourceByQuickID(resourceData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) executePostAction(w http.ResponseWriter, r *http.Request) {
	actions := make([]webHandlerTY.ActionConfig, 0)

	err := handlerUtils.LoadEntity(w, r, &actions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(actions) == 0 {
		http.Error(w, "there is no action supplied", http.StatusInternalServerError)
		return
	}

	for _, axn := range actions {
		resourceData := &handlerTY.ResourceData{
			QuickID: axn.Resource,
			KeyPath: axn.KayPath,
			Payload: axn.Payload,
		}
		err := h.action.ExecuteActionOnResourceByQuickID(resourceData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

}

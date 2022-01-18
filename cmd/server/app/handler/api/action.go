package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	"github.com/mycontroller-org/server/v2/pkg/api/action"
	webHandlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
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
func RegisterActionRoutes(router *mux.Router) {
	router.HandleFunc("/api/action", executeGetAction).Methods(http.MethodGet)
	router.HandleFunc("/api/action", executePostAction).Methods(http.MethodPost)
	router.HandleFunc("/api/action/node", executeNodeAction).Methods(http.MethodGet)
	router.HandleFunc("/api/action/gateway", executeGatewayAction).Methods(http.MethodGet)
}

func executeNodeAction(w http.ResponseWriter, r *http.Request) {
	query, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	actionArr := query[keyAction]
	idsArr := query[keyID]

	if len(actionArr) == 0 || len(idsArr) == 0 {
		http.Error(w, "required field(s) missing", 400)
		return
	}

	err = action.ExecuteNodeAction(actionArr[0], idsArr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func executeGatewayAction(w http.ResponseWriter, r *http.Request) {
	query, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	actionArr := query[keyAction]
	idsArr := query[keyID]

	if len(actionArr) == 0 || len(idsArr) == 0 {
		http.Error(w, "required field(s) missing", 400)
		return
	}

	err = action.ExecuteGatewayAction(actionArr[0], idsArr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func executeGetAction(w http.ResponseWriter, r *http.Request) {
	query, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	resourceArr := query[keyResource]
	keyPathArr := query[keyKeyPath]
	payloadArr := query[keyPayload]

	if len(resourceArr) == 0 || len(payloadArr) == 0 {
		http.Error(w, "required field(s) missing", 500)
		return
	}

	resourceData := &handlerTY.ResourceData{
		QuickID: resourceArr[0],
		Payload: payloadArr[0],
	}
	if len(keyPathArr) > 0 {
		resourceData.KeyPath = keyPathArr[0]
	}
	err = action.ExecuteActionOnResourceByQuickID(resourceData)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

func executePostAction(w http.ResponseWriter, r *http.Request) {
	actions := make([]webHandlerTY.ActionConfig, 0)

	err := handlerUtils.LoadEntity(w, r, actions)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if len(actions) == 0 {
		http.Error(w, "there is no action supplied", 500)
		return
	}

	for _, axn := range actions {
		resourceData := &handlerTY.ResourceData{
			QuickID: axn.Resource,
			KeyPath: axn.KayPath,
			Payload: axn.Payload,
		}
		err := action.ExecuteActionOnResourceByQuickID(resourceData)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

}

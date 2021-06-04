package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/backend/v2/cmd/core/app/handler/utils"
	"github.com/mycontroller-org/backend/v2/pkg/api/action"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
)

const (
	keyResource = "resource"
	keyPayload  = "payload"
	keySelector = "selector"
	keyAction   = "action"
	keyID       = "id"
)

// RegisterActionRoutes registers action api
func RegisterActionRoutes(router *mux.Router) {
	router.HandleFunc("/api/action", executeAction).Methods(http.MethodGet)
	router.HandleFunc("/api/action/node", executeNodeAction).Methods(http.MethodGet)
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

func executeAction(w http.ResponseWriter, r *http.Request) {
	query, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	resourceArr := query[keyResource]
	selectorArr := query[keySelector]
	payloadArr := query[keyPayload]

	if len(resourceArr) == 0 || len(payloadArr) == 0 {
		http.Error(w, "required field(s) missing", 500)
		return
	}

	resourceData := &handlerML.ResourceData{
		QuickID: resourceArr[0],
		Payload: payloadArr[0],
	}
	if len(selectorArr) > 0 {
		resourceData.Selector = selectorArr[0]
	}
	err = action.ExecuteActionOnResourceByQuickID(resourceData)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

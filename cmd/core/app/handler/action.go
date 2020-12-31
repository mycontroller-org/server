package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mycontroller-org/backend/v2/pkg/api/action"
)

const (
	keyResource = "resource"
	keyPayload  = "payload"
)

func registerActionRoutes(router *mux.Router) {
	router.HandleFunc("/api/action", executeAction).Methods(http.MethodGet)
}

func executeAction(w http.ResponseWriter, r *http.Request) {
	query, err := ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	resourceArr := query[keyResource]
	payloadArr := query[keyPayload]

	if len(resourceArr) == 0 || len(payloadArr) == 0 {
		http.Error(w, "required field(s) missing", 500)
		return
	}

	err = action.Execute(resourceArr[0], payloadArr[0])
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

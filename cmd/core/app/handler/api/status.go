package handler

import (
	"net/http"

	statusAPI "github.com/mycontroller-org/backend/v2/pkg/api/status"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/backend/v2/cmd/core/app/handler/utils"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/version"
)

// RegisterStatusRoutes registers status,version api
func RegisterStatusRoutes(router *mux.Router) {
	router.HandleFunc("/api/version", versionData).Methods(http.MethodGet)
	router.HandleFunc("/api/status", status).Methods(http.MethodGet)
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := statusAPI.GetMinimal()
	od, err := json.Marshal(&status)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	handlerUtils.WriteResponse(w, od)
}

func versionData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	v := version.Get()
	od, err := json.Marshal(&v)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	handlerUtils.WriteResponse(w, od)
}

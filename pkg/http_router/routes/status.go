package routes

import (
	"net/http"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
	"github.com/mycontroller-org/server/v2/pkg/version"
)

// RegisterStatusRoutes registers status,version api
func (h *Routes) registerStatusRoutes() {
	h.router.HandleFunc("/api/version", h.versionData).Methods(http.MethodGet)
	h.router.HandleFunc("/api/status", h.status).Methods(http.MethodGet)
}

func (h *Routes) status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := h.api.Status().GetMinimal()
	od, err := json.Marshal(&status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	handlerUtils.WriteResponse(w, od)
}

func (h *Routes) versionData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	v := version.Get()
	od, err := json.Marshal(&v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	handlerUtils.WriteResponse(w, od)
}

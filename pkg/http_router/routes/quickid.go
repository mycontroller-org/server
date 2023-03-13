package routes

import (
	"net/http"

	"github.com/mycontroller-org/server/v2/pkg/json"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
)

// RegisterQuickIDRoutes registers quickId api
func (h *Routes) registerQuickIDRoutes() {
	h.router.HandleFunc("/api/quickid", h.getResources).Methods(http.MethodGet)
}

func (h *Routes) getResources(w http.ResponseWriter, r *http.Request) {
	query, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ids := query["id"]
	if len(ids) == 0 {
		http.Error(w, "there is no id supplied", http.StatusBadRequest)
		return
	}

	result, err := h.quickIdAPI.GetResources(ids)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	handlerUtils.WriteResponse(w, data)
}

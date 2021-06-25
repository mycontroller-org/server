package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	quickIdAPI "github.com/mycontroller-org/server/v2/pkg/api/quickid"
	"github.com/mycontroller-org/server/v2/pkg/json"
)

// RegisterQuickIDRoutes registers quickId api
func RegisterQuickIDRoutes(router *mux.Router) {
	router.HandleFunc("/api/quickid", getResources).Methods(http.MethodGet)
}

func getResources(w http.ResponseWriter, r *http.Request) {
	query, err := handlerUtils.ReceivedQueryMap(r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	ids := query["id"]
	if len(ids) == 0 {
		http.Error(w, "there is no id supplied", 400)
		return
	}

	result, err := quickIdAPI.GetResources(ids)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	handlerUtils.WriteResponse(w, data)
}

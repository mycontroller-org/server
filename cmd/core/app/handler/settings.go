package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	settingsAPI "github.com/mycontroller-org/backend/v2/pkg/api/settings"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	settingsML "github.com/mycontroller-org/backend/v2/pkg/model/settings"
)

func registerSystemRoutes(router *mux.Router) {
	router.HandleFunc("/api/settings/system", getSystemSettings).Methods(http.MethodGet)
	router.HandleFunc("/api/settings/system", updateSettings).Methods(http.MethodPost)
}

func getSystemSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, err := settingsAPI.GetByID(settingsML.KeySystemSettings)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	od, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	WriteResponse(w, od)
}

func updateSettings(w http.ResponseWriter, r *http.Request) {
	entity := &settingsML.Settings{}
	err := LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be an empty", 400)
		return
	}
	err = settingsAPI.UpdateSettings(entity)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

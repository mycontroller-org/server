package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/server/v2/cmd/server/app/handler/utils"
	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	json "github.com/mycontroller-org/server/v2/pkg/json"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
)

// RegisterSystemRoutes registers settings api
func RegisterSystemRoutes(router *mux.Router) {
	router.HandleFunc("/api/settings", updateSettings).Methods(http.MethodPost)
	router.HandleFunc("/api/settings/system", getSystemSettings).Methods(http.MethodGet)
	router.HandleFunc("/api/settings/system/jwtsecret/reset", resetJwtSecret).Methods(http.MethodGet)
	router.HandleFunc("/api/settings/backuplocations", getSystemBackupLocations).Methods(http.MethodGet)
}

func resetJwtSecret(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := settingsAPI.ResetJwtAccessSecret("")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getSystemBackupLocations(w http.ResponseWriter, r *http.Request) {
	getSettings(settingsTY.KeySystemBackupLocations, w, r)
}

func getSystemSettings(w http.ResponseWriter, r *http.Request) {
	getSettings(settingsTY.KeySystemSettings, w, r)
}

func getSettings(key string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, err := settingsAPI.GetByID(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	od, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	handlerUtils.WriteResponse(w, od)
}

func updateSettings(w http.ResponseWriter, r *http.Request) {
	entity := &settingsTY.Settings{}
	err := handlerUtils.LoadEntity(w, r, entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entity.ID == "" {
		http.Error(w, "id should not be an empty", http.StatusBadRequest)
		return
	}
	err = settingsAPI.UpdateSettings(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

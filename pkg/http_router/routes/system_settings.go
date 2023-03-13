package routes

import (
	"net/http"

	json "github.com/mycontroller-org/server/v2/pkg/json"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
	handlerUtils "github.com/mycontroller-org/server/v2/pkg/utils/http_handler"
)

// RegisterSystemRoutes registers settings api
func (h *Routes) registerSystemRoutes() {
	h.router.HandleFunc("/api/settings", h.updateSettings).Methods(http.MethodPost)
	h.router.HandleFunc("/api/settings/system", h.getSystemSettings).Methods(http.MethodGet)
	h.router.HandleFunc("/api/settings/system/jwtsecret/reset", h.resetJwtSecret).Methods(http.MethodGet)
	h.router.HandleFunc("/api/settings/backuplocations", h.getSystemBackupLocations).Methods(http.MethodGet)
}

func (h *Routes) resetJwtSecret(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	err := h.api.Settings().ResetJwtAccessSecret("")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Routes) getSystemBackupLocations(w http.ResponseWriter, r *http.Request) {
	h.getSettings(settingsTY.KeySystemBackupLocations, w, r)
}

func (h *Routes) getSystemSettings(w http.ResponseWriter, r *http.Request) {
	h.getSettings(settingsTY.KeySystemSettings, w, r)
}

func (h *Routes) getSettings(key string, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, err := h.api.Settings().GetByID(key)
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

func (h *Routes) updateSettings(w http.ResponseWriter, r *http.Request) {
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
	err = h.api.Settings().UpdateSettings(entity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

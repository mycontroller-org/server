package handler

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	handlerUtils "github.com/mycontroller-org/backend/v2/cmd/core/app/handler/utils"
	settingsAPI "github.com/mycontroller-org/backend/v2/pkg/api/settings"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	settingsML "github.com/mycontroller-org/backend/v2/pkg/model/settings"
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/pkg/version"
)

// RegisterStatusRoutes registers status,version api
func RegisterStatusRoutes(router *mux.Router) {
	router.HandleFunc("/api/version", versionData).Methods(http.MethodGet)
	router.HandleFunc("/api/status", status).Methods(http.MethodGet)
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s := map[string]interface{}{
		"time": time.Now(),
	}
	hn, err := os.Hostname()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s["hostname"] = hn

	// include documentation url
	s["documentationUrl"] = cfg.CFG.Web.DocumentationURL

	// include login message
	rawSettings, err := settingsAPI.GetByID(settingsML.KeySystemSettings)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	sysSettings := &settingsML.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, rawSettings.Spec, sysSettings)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s["login"] = sysSettings.Login

	od, err := json.Marshal(&s)
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

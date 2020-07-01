package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/mycontroller-org/mycontroller/pkg/version"
)

func registerStatusRoutes(router *mux.Router) {
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
	}
	s["hostname"] = hn
	od, err := json.Marshal(&s)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(od)
}

func versionData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	v := version.Get()
	od, err := json.Marshal(&v)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write(od)
}

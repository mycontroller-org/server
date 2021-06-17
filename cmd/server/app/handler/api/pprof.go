package handler

import (
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
)

// RegisterPProfRoutes registers pprof api
func RegisterPProfRoutes(router *mux.Router) {
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline).Methods(http.MethodGet)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile).Methods(http.MethodGet)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol).Methods(http.MethodGet)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace).Methods(http.MethodGet)
	router.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index).Methods(http.MethodGet)
}

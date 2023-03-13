package routes

import (
	"net/http"
	"net/http/pprof"
)

// RegisterPProfRoutes registers pprof api
func (h *Routes) registerPProfRoutes() {
	h.router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline).Methods(http.MethodGet)
	h.router.HandleFunc("/debug/pprof/profile", pprof.Profile).Methods(http.MethodGet)
	h.router.HandleFunc("/debug/pprof/symbol", pprof.Symbol).Methods(http.MethodGet)
	h.router.HandleFunc("/debug/pprof/trace", pprof.Trace).Methods(http.MethodGet)
	h.router.PathPrefix("/debug/pprof/").HandlerFunc(pprof.Index).Methods(http.MethodGet)
}

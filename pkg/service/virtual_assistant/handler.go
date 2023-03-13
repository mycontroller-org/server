package service

import (
	"net/http"
	"strings"
)

const (
	ROOT_PATH = "/api/bot/virtual/assistant/"
)

// virtual assistant root route path
func (svc *VirtualAssistantService) registerServiceRoute() {
	svc.router.PathPrefix(ROOT_PATH).HandlerFunc(svc.handleRoute)
}

func (svc *VirtualAssistantService) handleRoute(w http.ResponseWriter, r *http.Request) {
	// get assistant id
	path := strings.TrimPrefix(r.URL.Path, ROOT_PATH)
	paths := strings.SplitN(path, "/", 2)

	assistantID := ""
	if len(paths) == 1 {
		r.URL.Path = ""
		assistantID = paths[0]
	} else if len(paths) == 2 {
		assistantID = paths[0]
		r.URL.Path = paths[1]
	}

	if assistantID == "" {
		http.Error(w, "assistant id can not be empty", http.StatusBadRequest)
		return
	}

	assistant := svc.store.Get(assistantID)
	if assistant == nil {
		http.Error(w, "requested assistant not available", http.StatusNotFound)
		return
	}

	assistant.ServeHTTP(w, r)
}

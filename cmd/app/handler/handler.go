package handler

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	svc "github.com/mycontroller-org/backend/v2/pkg/service"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

// StartHandler for http access
func StartHandler() error {
	router := mux.NewRouter()

	cfg := svc.CFG.Web

	// Enable Profiling, if enabled
	if cfg.EnableProfiling {
		registerPProfRoutes(router)
	}

	// register routes
	registerStatusRoutes(router)
	registerGatewayRoutes(router)
	registerNodeRoutes(router)
	registerSensorRoutes(router)
	registerSensorFieldRoutes(router)
	registerFirmwareRoutes(router)

	if cfg.WebDirectory != "" {
		fs := http.FileServer(http.Dir(cfg.WebDirectory))
		//router.Handle("/", fs)
		router.PathPrefix("/").Handler(fs)
	} else {
		defaultPage := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("Web directory not configured."))
		}
		router.HandleFunc("/", defaultPage)
	}

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})

	// Insert the middleware
	handler := c.Handler(router)

	addr := fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port)

	zap.L().Info("Listening HTTP service on", zap.String("address", addr), zap.String("webDirectory", cfg.WebDirectory))
	return http.ListenAndServe(addr, handler)
}

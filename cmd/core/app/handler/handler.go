package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	handlerML "github.com/mycontroller-org/backend/v2/pkg/model/handler"
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

// StartHandler for http access
func StartHandler() error {
	router := mux.NewRouter()

	cfg := cfg.CFG.Web

	// set JWT access secret in environment
	// TODO: this should be updated dynamically
	os.Setenv(handlerML.EnvJwtAccessSecret, "add2a90d-c7c5-4d93-96e2-e70eca62400d")

	// Enable Profiling, if enabled
	if cfg.EnableProfiling {
		registerPProfRoutes(router)
	}

	// register routes
	registerAuthRoutes(router)
	registerStatusRoutes(router)
	registerGatewayRoutes(router)
	registerNodeRoutes(router)
	registerSensorRoutes(router)
	registerSensorFieldRoutes(router)
	registerFirmwareRoutes(router)
	registerMetricRoutes(router)
	registerWebsocketRoutes(router)
	registerActionRoutes(router)
	registerDashboardRoutes(router)
	registerForwardPayloadRoutes(router)
	registerTaskRoutes(router)
	registerNotifyHandlerRoutes(router)
	registerSchedulerRoutes(router)
	registerDataRepositoryRoutes(router)
	registerSystemRoutes(router)

	if cfg.WebDirectory != "" {
		fs := http.FileServer(http.Dir(cfg.WebDirectory))
		//router.Handle("/", fs)
		router.PathPrefix("/").Handler(fs)
	} else {
		defaultPage := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			WriteResponse(w, []byte("Web directory not configured."))
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
	handler = middlewareAuthenticationVerification(handler)

	addr := fmt.Sprintf("%s:%d", cfg.BindAddress, cfg.Port)

	zap.L().Info("listening HTTP service on", zap.String("address", addr), zap.String("webDirectory", cfg.WebDirectory))
	return http.ListenAndServe(addr, handler)
}

func postErrorResponse(w http.ResponseWriter, message string, code int) {
	response := &handlerML.Response{
		Success: false,
		Message: message,
	}
	out, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	http.Error(w, string(out), code)
}

func postSuccessResponse(w http.ResponseWriter, data interface{}) {
	out, err := json.Marshal(data)
	if err != nil {
		postErrorResponse(w, err.Error(), 500)
		return
	}

	WriteResponse(w, out)
}

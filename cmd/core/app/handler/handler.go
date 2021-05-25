package handler

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model/config"
	webHandlerML "github.com/mycontroller-org/backend/v2/pkg/model/web_handler"
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
	mcWS "github.com/mycontroller-org/backend/v2/pkg/service/websocket"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

// GetHandler for http access
func GetHandler() (http.Handler, error) {
	router := mux.NewRouter()

	webCfg := cfg.CFG.Web

	// set JWT access secret in environment
	// TODO: this should be updated dynamically
	os.Setenv(webHandlerML.EnvJwtAccessSecret, "add2a90d-c7c5-4d93-96e2-e70eca62400d")

	// Enable Profiling, if enabled
	if webCfg.EnableProfiling {
		registerPProfRoutes(router)
	}

	// register routes
	registerAuthRoutes(router)
	registerStatusRoutes(router)
	mcWS.RegisterWebsocketRoutes(router)
	registerGatewayRoutes(router)
	registerNodeRoutes(router)
	registerSourceRoutes(router)
	registerFieldRoutes(router)
	registerFirmwareRoutes(router)
	registerMetricRoutes(router)
	registerActionRoutes(router)
	registerDashboardRoutes(router)
	registerForwardPayloadRoutes(router)
	registerTaskRoutes(router)
	registerNotifyHandlerRoutes(router)
	registerSchedulerRoutes(router)
	registerDataRepositoryRoutes(router)
	registerSystemRoutes(router)
	registerQuickIDRoutes(router)
	registerImportExportRoutes(router)

	// add secure and insecure directories into handler
	addFileServers(cfg.CFG.Directories, router)

	if webCfg.WebDirectory != "" {
		fs := http.FileServer(http.Dir(webCfg.WebDirectory))
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

	return handler, nil
}

func postErrorResponse(w http.ResponseWriter, message string, code int) {
	response := &webHandlerML.Response{
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

// configure secure and insecure dir shares
func addFileServers(dirs config.Directories, router *mux.Router) {
	if dirs.SecureShare != "" {
		zap.L().Info("secure share directory included", zap.String("directory", dirs.SecureShare), zap.String("handlerPath", webHandlerML.SecureShareDirWebHandlerPath))
		fs := http.StripPrefix(webHandlerML.SecureShareDirWebHandlerPath, http.FileServer(http.Dir(dirs.SecureShare)))
		router.PathPrefix(webHandlerML.SecureShareDirWebHandlerPath).Handler(fs)
	}
	if dirs.InsecureShare != "" {
		zap.L().Info("insecure share directory included", zap.String("directory", dirs.InsecureShare), zap.String("handlerPath", webHandlerML.InsecureShareDirWebHandlerPath))
		fs := http.StripPrefix(webHandlerML.InsecureShareDirWebHandlerPath, http.FileServer(http.Dir(dirs.InsecureShare)))
		router.PathPrefix(webHandlerML.InsecureShareDirWebHandlerPath).Handler(fs)

	}
}

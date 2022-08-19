package handler

import (
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	handlerAPI "github.com/mycontroller-org/server/v2/cmd/server/app/handler/api"
	handlerAuthAPI "github.com/mycontroller-org/server/v2/cmd/server/app/handler/api/auth"
	middleware "github.com/mycontroller-org/server/v2/cmd/server/app/handler/middleware"
	webConsole "github.com/mycontroller-org/server/v2/cmd/server/app/web-console"
	virtualAssistantAPI "github.com/mycontroller-org/server/v2/pkg/service/virtual_assistant"
	mcWS "github.com/mycontroller-org/server/v2/pkg/service/websocket"
	"github.com/mycontroller-org/server/v2/pkg/store"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	webHandlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

// GetHandler for http access
func GetHandler() (http.Handler, error) {
	router := mux.NewRouter()

	webCfg := store.CFG.Web

	// set JWT access secret in environment
	// TODO: this should be updated dynamically
	if os.Getenv(webHandlerTY.EnvJwtAccessSecret) == "" {
		os.Setenv(webHandlerTY.EnvJwtAccessSecret, "add2a90d-c7c5-4d93-96e2-e70eca62400d")
	}

	// Enable Profiling, if enabled
	if webCfg.EnableProfiling {
		handlerAPI.RegisterPProfRoutes(router)
	}

	// authentication routes
	handlerAuthAPI.RegisterAuthRoutes(router)
	handlerAuthAPI.RegisterOAuthRoutes(router)

	// other routes
	handlerAPI.RegisterStatusRoutes(router)
	handlerAPI.RegisterGatewayRoutes(router)
	handlerAPI.RegisterNodeRoutes(router)
	handlerAPI.RegisterSourceRoutes(router)
	handlerAPI.RegisterFieldRoutes(router)
	handlerAPI.RegisterFirmwareRoutes(router)
	handlerAPI.RegisterMetricRoutes(router)
	handlerAPI.RegisterActionRoutes(router)
	handlerAPI.RegisterDashboardRoutes(router)
	handlerAPI.RegisterForwardPayloadRoutes(router)
	handlerAPI.RegisterTaskRoutes(router)
	handlerAPI.RegisterHandlerRoutes(router)
	handlerAPI.RegisterSchedulerRoutes(router)
	handlerAPI.RegisterDataRepositoryRoutes(router)
	handlerAPI.RegisterSystemRoutes(router)
	handlerAPI.RegisterQuickIDRoutes(router)
	handlerAPI.RegisterBackupRestoreRoutes(router)
	handlerAPI.RegisterVirtualDeviceRoutes(router)
	handlerAPI.RegisterVirtualAssistantRoutes(router)

	// virtual assistants service route
	virtualAssistantAPI.RegisterVirtualAssistantServiceRoutes(router)

	// websocket router
	mcWS.RegisterWebsocketRoutes(router)

	// add secure and insecure directories into handler
	addFileServers(store.CFG.Directories, router)

	if webCfg.WebDirectory != "" {
		fs := http.FileServer(http.Dir(webCfg.WebDirectory))
		//router.Handle("/", fs)
		router.PathPrefix("/").Handler(fs)
	} else {
		//defaultPage := func(w http.ResponseWriter, r *http.Request) {
		//	w.Header().Set("Content-Type", "text/plain")
		//	handlerUtils.WriteResponse(w, []byte("Web directory not configured."))
		//}
		// router.HandleFunc("/", defaultPage)
		fs := http.FileServer(webConsole.StaticFiles)
		router.PathPrefix("/").Handler(fs)
	}

	// pre flight middleware
	withCors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		// Enable Debugging for testing, consider disabling in production
		Debug: false,
	})
	withPreflight := withCors.Handler(router)

	// include authentication middleware
	withAuthentication := middleware.MiddlewareAuthenticationVerification(withPreflight)

	// include gzip middleware
	withGzip := gziphandler.GzipHandler(withAuthentication)

	return withGzip, nil
}

// configure secure and insecure dir shares
func addFileServers(dirs config.Directories, router *mux.Router) {
	if dirs.SecureShare != "" {
		zap.L().Info("secure share directory included", zap.String("directory", dirs.SecureShare), zap.String("handlerPath", webHandlerTY.SecureShareDirWebHandlerPath))
		fs := http.StripPrefix(webHandlerTY.SecureShareDirWebHandlerPath, http.FileServer(http.Dir(dirs.SecureShare)))
		router.PathPrefix(webHandlerTY.SecureShareDirWebHandlerPath).Handler(fs)
	}
	if dirs.InsecureShare != "" {
		zap.L().Info("insecure share directory included", zap.String("directory", dirs.InsecureShare), zap.String("handlerPath", webHandlerTY.InsecureShareDirWebHandlerPath))
		fs := http.StripPrefix(webHandlerTY.InsecureShareDirWebHandlerPath, http.FileServer(http.Dir(dirs.InsecureShare)))
		router.PathPrefix(webHandlerTY.InsecureShareDirWebHandlerPath).Handler(fs)

	}
}

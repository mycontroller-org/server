package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	entitiesAPI "github.com/mycontroller-org/server/v2/pkg/api/entities"
	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	encryptionAPI "github.com/mycontroller-org/server/v2/pkg/encryption"
	middleware "github.com/mycontroller-org/server/v2/pkg/http_router/middleware"
	routes "github.com/mycontroller-org/server/v2/pkg/http_router/routes"
	authRoutes "github.com/mycontroller-org/server/v2/pkg/http_router/routes/auth"
	webConsole "github.com/mycontroller-org/server/v2/pkg/http_router/web-console"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	webHandlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
	loggerUtils "github.com/mycontroller-org/server/v2/pkg/utils/logger"
	busTY "github.com/mycontroller-org/server/v2/plugin/bus/types"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"github.com/rs/cors"
	"go.uber.org/zap"
)

// NewHandler for http access
func New(ctx context.Context, cfg *config.Config, router *mux.Router) (http.Handler, error) {
	webCfg := cfg.Web
	logger, err := loggerUtils.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	storage, err := storageTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	bus, err := busTY.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	enc, err := encryptionAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	coreApi, err := entitiesAPI.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	namedLogger := logger.Named("route_handler")

	// set JWT access secret in environment
	_settingsAPI := settingsAPI.New(ctx, logger, storage, enc, bus)
	err = _settingsAPI.UpdateJwtAccessSecret()
	if err != nil {
		namedLogger.Error("error on getting jwt access secret", zap.Error(err))
		return nil, err
	}

	// register application api routes
	_, err = routes.New(ctx, router, webCfg.EnableProfiling)
	if err != nil {
		return nil, err
	}

	// register authentication routes, used in google, alexa and others
	_authRoutes := authRoutes.NewAuthRoutes(logger, coreApi, router)
	_oAuthRoutes := authRoutes.NewOAuthRoutes(logger, coreApi, router)
	_authRoutes.RegisterRoutes()
	_oAuthRoutes.RegisterRoutes()

	// add secure and insecure directories into handler
	addFileServers(namedLogger, cfg.Directories, router)

	if webCfg.WebDirectory != "" {
		fs := http.FileServer(http.Dir(webCfg.WebDirectory))
		router.PathPrefix("/").Handler(fs)
	} else {
		//defaultPage := func(w http.ResponseWriter, r *http.Request) {
		//	w.Header().Set("Content-Type", "text/plain")
		//	handlerUtils.WriteResponse(w, []byte("Web directory not configured."))
		//}
		// router.HandleFunc("/", defaultPage)

		// embedded static files
		fs := http.FileServer(webConsole.StaticFiles)
		router.PathPrefix("/").Handler(fs)
	}

	// pre flight middleware
	withCors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		MaxAge:         int(time.Hour * 24 / time.Second), // 24 hours
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
func addFileServers(logger *zap.Logger, dirs config.Directories, router *mux.Router) {
	if dirs.SecureShare != "" {
		logger.Info("secure share directory included", zap.String("directory", dirs.SecureShare), zap.String("handlerPath", webHandlerTY.SecureShareDirWebHandlerPath))
		fs := http.StripPrefix(webHandlerTY.SecureShareDirWebHandlerPath, http.FileServer(http.Dir(dirs.SecureShare)))
		router.PathPrefix(webHandlerTY.SecureShareDirWebHandlerPath).Handler(fs)
	}
	if dirs.InsecureShare != "" {
		logger.Info("insecure share directory included", zap.String("directory", dirs.InsecureShare), zap.String("handlerPath", webHandlerTY.InsecureShareDirWebHandlerPath))
		fs := http.StripPrefix(webHandlerTY.InsecureShareDirWebHandlerPath, http.FileServer(http.Dir(dirs.InsecureShare)))
		router.PathPrefix(webHandlerTY.InsecureShareDirWebHandlerPath).Handler(fs)
	}
}

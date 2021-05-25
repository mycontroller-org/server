package mainhandler

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/mycontroller-org/backend/v2/cmd/core/app/handler"
	"github.com/mycontroller-org/backend/v2/cmd/core/https"
	cfg "github.com/mycontroller-org/backend/v2/pkg/service/configuration"
)

func StartHandler() {
	webCfg := cfg.CFG.Web

	if !webCfg.HTTP.Enabled && !webCfg.SSL.Enabled && !webCfg.Letsencrypt.Enabled {
		zap.L().Fatal("web services are disabled. Enable at least a service HTTP, HTTPS/SSL or HTTPS/Letsencrypt")
	}

	handler, err := handler.GetHandler()
	if err != nil {
		zap.L().Fatal("Error on getting handler", zap.Error(err))
	}

	zap.L().Info("web console direcory location", zap.String("web_directory", webCfg.WebDirectory))

	errs := make(chan error, 1) // a channel for errors

	// http service
	if webCfg.HTTP.Enabled {
		go func() {
			addr := fmt.Sprintf("%s:%d", webCfg.HTTP.BindAddress, webCfg.HTTP.Port)
			zap.L().Info("listening HTTP service on", zap.String("address", addr))
			server := &http.Server{
				Addr:    addr,
				Handler: handler,
			}

			err = server.ListenAndServe()
			if err != nil {
				zap.L().Error("Error on starting http handler", zap.Error(err))
				errs <- err
			}
		}()
	}

	// ssl service
	if webCfg.SSL.Enabled {
		go func() {
			addr := fmt.Sprintf("%s:%d", webCfg.SSL.BindAddress, webCfg.SSL.Port)
			zap.L().Info("listening HTTPS/SSL service on", zap.String("address", addr))

			tlsConfig, err := https.GetSSLTLSConfig(webCfg.SSL)
			if err != nil {
				zap.L().Error("error on getting https/ssl tlsConfig", zap.Error(err), zap.Any("sslConfig", webCfg.SSL))
				errs <- err
				return
			}

			server := &http.Server{
				Addr:      addr,
				TLSConfig: tlsConfig,
				Handler:   handler,
			}

			err = server.ListenAndServeTLS("", "")
			if err != nil {
				zap.L().Error("Error on starting https/ssl handler", zap.Error(err))
				errs <- err
			}
		}()
	}

	// letsencrypt service
	if webCfg.Letsencrypt.Enabled {
		go func() {
			addr := fmt.Sprintf("%s:%d", webCfg.Letsencrypt.BindAddress, webCfg.Letsencrypt.Port)
			zap.L().Info("listening HTTPS/Letsencrypt service on", zap.String("address", addr))

			tlsConfig, err := https.GetLetsencryptTLSConfig(webCfg.Letsencrypt)
			if err != nil {
				zap.L().Error("error on getting letsencrypt tlsConfig", zap.Error(err), zap.Any("letsencryptConfig", webCfg.Letsencrypt))
				errs <- err
				return
			}

			server := &http.Server{
				Addr:      addr,
				TLSConfig: tlsConfig,
				Handler:   handler,
			}

			err = server.ListenAndServeTLS("", "")
			if err != nil {
				zap.L().Error("Error on starting https/letsencrypt handler", zap.Error(err))
				errs <- err
			}
		}()
	}

	// This will run forever until channel receives error
	err = <-errs
	zap.L().Fatal("Error on starting a handler service", zap.Error(err))
}

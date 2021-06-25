package listener

import (
	"fmt"
	"log"
	"net/http"

	"go.uber.org/zap"

	"github.com/mycontroller-org/server/v2/cmd/server/https"
	cfg "github.com/mycontroller-org/server/v2/pkg/service/configuration"
)

const (
	LoggerPrefixHTTP = "HTTP"
	LoggerPrefixSSL  = "HTTPS/SSL"
	LoggerPrefixACME = "HTTPS/ACME"
)

func StartListener(handler http.Handler) {
	loggerCfg := cfg.CFG.Logger
	webCfg := cfg.CFG.Web

	if !webCfg.Http.Enabled && !webCfg.HttpsSSL.Enabled && !webCfg.HttpsACME.Enabled {
		zap.L().Fatal("web services are disabled. Enable at least a service HTTP, HTTPS/SSL or HTTPS/ACME")
	}

	zap.L().Info("web console direcory location", zap.String("web_directory", webCfg.WebDirectory))

	errs := make(chan error, 1) // a channel for errors

	// http service
	if webCfg.Http.Enabled {
		go func() {
			addr := fmt.Sprintf("%s:%d", webCfg.Http.BindAddress, webCfg.Http.Port)
			zap.L().Info("listening HTTP service on", zap.String("address", addr))
			server := &http.Server{
				Addr:     addr,
				Handler:  handler,
				ErrorLog: log.New(getLogger(LoggerPrefixHTTP, loggerCfg.Mode, loggerCfg.Level.WebHandler, loggerCfg.Encoding), "", 0),
			}

			err := server.ListenAndServe()
			if err != nil {
				zap.L().Error("Error on starting http handler", zap.Error(err))
				errs <- err
			}
		}()
	}

	// https ssl service
	if webCfg.HttpsSSL.Enabled {
		go func() {
			addr := fmt.Sprintf("%s:%d", webCfg.HttpsSSL.BindAddress, webCfg.HttpsSSL.Port)
			zap.L().Info("listening HTTPS/SSL service on", zap.String("address", addr))

			tlsConfig, err := https.GetSSLTLSConfig(webCfg.HttpsSSL)
			if err != nil {
				zap.L().Error("error on getting https/ssl tlsConfig", zap.Error(err), zap.Any("sslConfig", webCfg.HttpsSSL))
				errs <- err
				return
			}

			server := &http.Server{
				Addr:      addr,
				TLSConfig: tlsConfig,
				Handler:   handler,
				ErrorLog:  log.New(getLogger(LoggerPrefixSSL, loggerCfg.Mode, loggerCfg.Level.WebHandler, loggerCfg.Encoding), "", 0),
			}

			err = server.ListenAndServeTLS("", "")
			if err != nil {
				zap.L().Error("Error on starting https/ssl handler", zap.Error(err))
				errs <- err
			}
		}()
	}

	// https acme service
	if webCfg.HttpsACME.Enabled {
		go func() {
			addr := fmt.Sprintf("%s:%d", webCfg.HttpsACME.BindAddress, webCfg.HttpsACME.Port)
			zap.L().Info("listening HTTPS/acme service on", zap.String("address", addr))

			tlsConfig, err := https.GetAcmeTLSConfig(webCfg.HttpsACME)
			if err != nil {
				zap.L().Error("error on getting acme tlsConfig", zap.Error(err), zap.Any("acmeConfig", webCfg.HttpsACME))
				errs <- err
				return
			}

			server := &http.Server{
				Addr:      addr,
				TLSConfig: tlsConfig,
				Handler:   handler,
				ErrorLog:  log.New(getLogger(LoggerPrefixACME, loggerCfg.Mode, loggerCfg.Level.WebHandler, loggerCfg.Encoding), "", 0),
			}

			err = server.ListenAndServeTLS("", "")
			if err != nil {
				zap.L().Error("Error on starting https/acme handler", zap.Error(err))
				errs <- err
			}
		}()
	}

	// This will run forever until channel receives error
	err := <-errs
	zap.L().Fatal("Error on starting a handler service", zap.Error(err))
}

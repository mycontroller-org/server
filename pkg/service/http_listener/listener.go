package listener

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/mycontroller-org/server/v2/pkg/service/http_listener/https"
	"github.com/mycontroller-org/server/v2/pkg/types/config"
	contextTY "github.com/mycontroller-org/server/v2/pkg/types/context"
	serviceTY "github.com/mycontroller-org/server/v2/pkg/types/service"
	"github.com/mycontroller-org/server/v2/pkg/utils"
)

const (
	LoggerPrefixHTTP = "HTTP"
	LoggerPrefixSSL  = "HTTPS/SSL"
	LoggerPrefixACME = "HTTPS/ACME"

	defaultReadTimeout = time.Second * 60
)

type HttpListener struct {
	logger  *zap.Logger
	config  config.WebConfig
	handler http.Handler
}

func New(ctx context.Context, cfg config.WebConfig, handler http.Handler) (serviceTY.Service, error) {
	logger, err := contextTY.LoggerFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return &HttpListener{
		logger:  logger.Named("http_listener"),
		config:  cfg,
		handler: handler,
	}, nil
}

func (svc *HttpListener) Name() string {
	return "http_listener_service"
}

func (l *HttpListener) Start() error {
	if !l.config.Http.Enabled && !l.config.HttpsSSL.Enabled && !l.config.HttpsACME.Enabled {
		l.logger.Fatal("web services are disabled. enable at least a service: HTTP, HTTPS/SSL or HTTPS/ACME")
	}

	l.logger.Info("web console directory location", zap.String("web_directory", l.config.WebDirectory))

	errs := make(chan error, 1) // a channel for errors

	// get readTimeout
	readTimeout := utils.ToDuration(l.config.ReadTimeout, defaultReadTimeout)
	l.logger.Debug("web server connection timeout", zap.String("read_timeout", readTimeout.String()))

	// http service
	if l.config.Http.Enabled {
		go func() {
			addr := fmt.Sprintf("%s:%d", l.config.Http.BindAddress, l.config.Http.Port)
			l.logger.Info("listening HTTP service on", zap.String("address", addr))
			server := &http.Server{
				ReadTimeout: readTimeout,
				Addr:        addr,
				Handler:     l.handler,
				ErrorLog:    log.New(getLogger(LoggerPrefixHTTP, l.logger), "", 0),
			}

			err := server.ListenAndServe()
			if err != nil {
				l.logger.Error("Error on starting http handler", zap.Error(err))
				errs <- err
			}
		}()
	}

	// https ssl service
	if l.config.HttpsSSL.Enabled {
		go func() {
			addr := fmt.Sprintf("%s:%d", l.config.HttpsSSL.BindAddress, l.config.HttpsSSL.Port)
			l.logger.Info("listening HTTPS/SSL service on", zap.String("address", addr))

			tlsConfig, err := https.GetSSLTLSConfig(l.logger, l.config.HttpsSSL)
			if err != nil {
				l.logger.Error("error on getting https/ssl tlsConfig", zap.Error(err), zap.Any("sslConfig", l.config.HttpsSSL))
				errs <- err
				return
			}

			server := &http.Server{
				ReadTimeout: readTimeout,
				Addr:        addr,
				TLSConfig:   tlsConfig,
				Handler:     l.handler,
				ErrorLog:    log.New(getLogger(LoggerPrefixSSL, l.logger), "", 0),
			}

			err = server.ListenAndServeTLS("", "")
			if err != nil {
				l.logger.Error("error on starting https/ssl handler", zap.Error(err))
				errs <- err
			}
		}()
	}

	// https acme service
	if l.config.HttpsACME.Enabled {
		go func() {
			addr := fmt.Sprintf("%s:%d", l.config.HttpsACME.BindAddress, l.config.HttpsACME.Port)
			l.logger.Info("listening HTTPS/acme service on", zap.String("address", addr))

			tlsConfig, err := https.GetAcmeTLSConfig(l.config.HttpsACME)
			if err != nil {
				l.logger.Error("error on getting acme tlsConfig", zap.Error(err), zap.Any("acmeConfig", l.config.HttpsACME))
				errs <- err
				return
			}

			server := &http.Server{
				ReadTimeout: readTimeout,
				Addr:        addr,
				TLSConfig:   tlsConfig,
				Handler:     l.handler,
				ErrorLog:    log.New(getLogger(LoggerPrefixACME, l.logger), "", 0),
			}

			err = server.ListenAndServeTLS("", "")
			if err != nil {
				l.logger.Error("error on starting https/acme handler", zap.Error(err))
				errs <- err
			}
		}()
	}

	// This will run forever until channel receives error
	err := <-errs
	l.logger.Fatal("error on starting a handler service", zap.Error(err))
	return err
}

func (l *HttpListener) Close() error {
	return nil
}

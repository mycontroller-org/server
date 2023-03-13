package natsio

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"

	"github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

// CustomDialer struct
type CustomDialer struct {
	uri    *url.URL
	config *Config
	logger *zap.Logger
}

// NewCustomDialer returns a custom dialer
func NewCustomDialer(cfg *Config, logger *zap.Logger) (*CustomDialer, error) {
	uri, err := url.ParseRequestURI(cfg.ServerURL)
	if err != nil {
		return nil, err
	}
	cd := &CustomDialer{
		uri:    uri,
		config: cfg,
		logger: logger,
	}
	return cd, nil
}

// Dial implementation
func (cd *CustomDialer) Dial(network, address string) (net.Conn, error) {
	cd.logger.Debug("connecting via custom dialer", zap.String("server", cd.uri.String()))

	timeout := utils.ToDuration(cd.config.ConnectionTimeout, defaultConnectionTimeout)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cd.config.Insecure,
	}
	switch cd.uri.Scheme {
	case "ws":
		return NewWebsocket(cd.uri.String(), nil, timeout, cd.config.WebsocketOptions)

	case "wss":
		return NewWebsocket(cd.uri.String(), tlsConfig, timeout, cd.config.WebsocketOptions)

	case "tcp", "http", "nats":
		conn, err := net.DialTimeout("tcp", cd.uri.Host, timeout)
		if err != nil {
			cd.logger.Debug("dialer error", zap.Error(err))
			return nil, err
		}
		return conn, nil

	case "tls", "nats+tls":
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", cd.uri.Host, tlsConfig)
		if err != nil {
			cd.logger.Debug("dialer error", zap.Error(err))
			return nil, err
		}
		return conn, nil
	}
	cd.logger.Debug("unknown protocol", zap.String("protocol", cd.uri.Scheme))
	return nil, fmt.Errorf("[BUS:NATS.IO] unknown protocol:%s", cd.uri.Scheme)
}

package natsio

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/url"

	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"go.uber.org/zap"
)

// CustomDialer struct
type CustomDialer struct {
	uri    *url.URL
	config *Config
}

// NewCustomDialer returns a custom dialer
func NewCustomDialer(cfg *Config) (*CustomDialer, error) {
	uri, err := url.ParseRequestURI(cfg.ServerURL)
	if err != nil {
		return nil, err
	}
	cd := &CustomDialer{
		uri:    uri,
		config: cfg,
	}
	return cd, nil
}

// Dial implementation
func (cd *CustomDialer) Dial(network, address string) (net.Conn, error) {
	PrintDebug("connecting via custom dialer", zap.String("server", cd.uri.String()))

	timeout := utils.ToDuration(cd.config.ConnectionTimeout, connectionTimeoutDefault)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cd.config.TLSInsecureSkipVerify,
	}
	switch cd.uri.Scheme {
	case "ws":
		return NewWebsocket(cd.uri.String(), nil, timeout, cd.config.WebsocketOptions)

	case "wss":
		return NewWebsocket(cd.uri.String(), tlsConfig, timeout, cd.config.WebsocketOptions)

	case "tcp", "http", "nats":
		conn, err := net.DialTimeout("tcp", cd.uri.Host, timeout)
		if err != nil {
			PrintError("dialer error", zap.Error(err))
			return nil, err
		}
		return conn, nil

	case "tls", "nats+tls":
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", cd.uri.Host, tlsConfig)
		if err != nil {
			PrintError("dialer error", zap.Error(err))
			return nil, err
		}
		return conn, nil
	}
	PrintError("unknown protocol", zap.String("protocol", cd.uri.Scheme))
	return nil, fmt.Errorf("[BUS:NATS.IO] unknown protocol:%s", cd.uri.Scheme)
}

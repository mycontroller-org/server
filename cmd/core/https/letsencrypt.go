package https

import (
	"crypto/tls"
	"errors"

	"github.com/mycontroller-org/backend/v2/pkg/model/config"
	"golang.org/x/crypto/acme/autocert"
)

// GetLetsencryptTLSConfig return letsencrypt certificate
func GetLetsencryptTLSConfig(cfg config.LetsencryptConfig) (*tls.Config, error) {
	if cfg.CacheDir == "" {
		return nil, errors.New("empty cache dir not allowed")
	}
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.Domains...),
		Email:      cfg.Email,
		Cache:      autocert.DirCache(cfg.CacheDir),
	}

	return certManager.TLSConfig(), nil
}

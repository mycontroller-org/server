package https

import (
	"crypto/tls"
	"errors"

	"github.com/mycontroller-org/server/v2/pkg/types/config"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

// GetAcmeTLSConfig return certificate from acme
func GetAcmeTLSConfig(cfg config.HttpsACMEConfig) (*tls.Config, error) {
	if cfg.CacheDir == "" {
		return nil, errors.New("empty cache dir not allowed")
	}

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(cfg.Domains...),
		Email:      cfg.Email,
		Cache:      autocert.DirCache(cfg.CacheDir),
	}

	// if acme directory supplied use the custom one
	if cfg.ACMEDirectory != "" {
		certManager.Client = &acme.Client{DirectoryURL: cfg.ACMEDirectory}
	}

	return certManager.TLSConfig(), nil
}

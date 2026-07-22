package https

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/config"
	schedulerTY "github.com/mycontroller-org/server/v2/pkg/types/scheduler"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

const (
	CustomCertFileName = "custom.crt"
	CustomKeyFileName  = "custom.key"

	RSABits               = 2048
	OrganizationName      = "MyController.org"
	ValidityDays          = 365 // days - default generated certificate lifetime
	RenewBeforeDays       = 30  // days - default regenerate when remaining validity is less than this
	GeneratedCertFileName = "mc_generated.crt"
	GeneratedKeyFileName  = "mc_generated.key"

	// daily renewal check for MyController-managed self-signed certificates
	sslRenewalJobName = "https_ssl_cert_renewal"
	sslRenewalCron    = "@every 24h"
)

// SSLManager loads HTTPS certificates and, when MyController manages them,
// periodically renews the cert (reusing the private key) and hot-reloads it.
type SSLManager struct {
	logger  *zap.Logger
	cfg     config.HttpsSSLConfig
	managed bool

	mu   sync.RWMutex
	cert *tls.Certificate

	scheduler schedulerTY.CoreScheduler
}

// NewSSLManager prepares certificates for HTTPS/SSL.
// Managed (auto-generated) certs are renewed on a daily schedule when a scheduler is provided.
// Custom certs (custom.crt + custom.key) are loaded as-is and never auto-renewed.
func NewSSLManager(logger *zap.Logger, cfg config.HttpsSSLConfig, scheduler schedulerTY.CoreScheduler) (*SSLManager, error) {
	if cfg.CertDir == "" {
		return nil, errors.New("cert_dir is missing")
	}

	m := &SSLManager{
		logger:    logger,
		cfg:       cfg,
		managed:   isManagedByMyController(cfg),
		scheduler: scheduler,
	}

	// only relevant for MyController-managed self-signed certificates
	if m.managed {
		validityDays := resolveValidityDays(cfg.ValidityDays)
		renewBeforeDays := resolveRenewBeforeDays(cfg.RenewBeforeDays)
		if validityDays <= renewBeforeDays {
			logger.Warn("invalid SSL self-signed certificate configuration: validity_days must be greater than renew_before_days",
				zap.Int("validityDays", validityDays), zap.Int("renewBeforeDays", renewBeforeDays),
				zap.String("hint", "increase validity_days or decrease renew_before_days; otherwise a newly generated certificate remains below the renewal threshold and will be regenerated on every check"),
			)
		}
	}

	if err := m.loadOrCreate(); err != nil {
		return nil, err
	}

	return m, nil
}

// TLSConfig returns a tls.Config that serves the current certificate and
// picks up renewals without restarting the listener.
func (m *SSLManager) TLSConfig() *tls.Config {
	return &tls.Config{
		GetCertificate: m.getCertificate,
	}
}

// Managed reports whether certificates are generated/renewed by MyController.
func (m *SSLManager) Managed() bool {
	return m.managed
}

// StartDailyRenewalCheck schedules a once-per-day renewal check when SSL is
// managed by MyController. No-op for custom certificates or when scheduler is nil.
func (m *SSLManager) StartDailyRenewalCheck() error {
	if !m.managed {
		m.logger.Debug("SSL certificate is custom; daily renewal check is disabled")
		return nil
	}
	if m.scheduler == nil {
		m.logger.Warn("core scheduler not available; SSL daily renewal check is disabled")
		return nil
	}

	err := m.scheduler.AddFunc(sslRenewalJobName, sslRenewalCron, m.dailyRenewalCheck)
	if err != nil {
		return fmt.Errorf("schedule SSL daily renewal check: %w", err)
	}
	m.logger.Info("scheduled daily SSL certificate renewal check", zap.String("job", sslRenewalJobName), zap.String("cron", sslRenewalCron),
		zap.Int("renewBeforeDays", resolveRenewBeforeDays(m.cfg.RenewBeforeDays)), zap.Int("validityDays", resolveValidityDays(m.cfg.ValidityDays)))
	return nil
}

// Close stops the daily renewal check if it was scheduled.
func (m *SSLManager) Close() error {
	if m.scheduler != nil && m.managed {
		m.scheduler.RemoveFunc(sslRenewalJobName)
	}
	return nil
}

// CheckAndRenew regenerates the managed certificate when remaining validity is
// below the configured threshold. Safe to call from the daily job or tests.
func (m *SSLManager) CheckAndRenew() error {
	if !m.managed {
		return nil
	}

	certFile := filepath.Join(m.cfg.CertDir, GeneratedCertFileName)
	keyFile := filepath.Join(m.cfg.CertDir, GeneratedKeyFileName)
	validityDays := resolveValidityDays(m.cfg.ValidityDays)
	renewBeforeDays := resolveRenewBeforeDays(m.cfg.RenewBeforeDays)

	if !shouldRegenerateCert(m.logger, certFile, keyFile, renewBeforeDays) {
		m.logger.Debug("SSL certificate still valid; no renewal needed", zap.String("certFile", certFile), zap.Int("renewBeforeDays", renewBeforeDays))
		return nil
	}

	if err := generateSSLCert(m.logger, m.cfg.CertDir, validityDays); err != nil {
		return err
	}

	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.cert = &certificate
	m.mu.Unlock()

	m.logger.Info("SSL certificate renewed and hot-reloaded")
	return nil
}

func (m *SSLManager) getCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.cert == nil {
		return nil, errors.New("SSL certificate not loaded")
	}
	return m.cert, nil
}

func (m *SSLManager) dailyRenewalCheck() {
	if err := m.CheckAndRenew(); err != nil {
		m.logger.Error("error on daily SSL certificate renewal check", zap.Error(err))
	}
}

func (m *SSLManager) loadOrCreate() error {
	var certFile, keyFile string

	if !m.managed {
		certFile = filepath.Join(m.cfg.CertDir, CustomCertFileName)
		keyFile = filepath.Join(m.cfg.CertDir, CustomKeyFileName)
		m.logger.Info("using custom SSL certificate", zap.String("certFile", certFile), zap.String("keyFile", keyFile))
	} else {
		certFile = filepath.Join(m.cfg.CertDir, GeneratedCertFileName)
		keyFile = filepath.Join(m.cfg.CertDir, GeneratedKeyFileName)
		validityDays := resolveValidityDays(m.cfg.ValidityDays)
		renewBeforeDays := resolveRenewBeforeDays(m.cfg.RenewBeforeDays)

		if shouldRegenerateCert(m.logger, certFile, keyFile, renewBeforeDays) {
			if err := generateSSLCert(m.logger, m.cfg.CertDir, validityDays); err != nil {
				return err
			}
		} else {
			m.logger.Info("using existing generated SSL certificate", zap.String("certFile", certFile), zap.String("keyFile", keyFile))
		}
	}

	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.cert = &certificate
	m.mu.Unlock()
	return nil
}

// isManagedByMyController is true when no operator-supplied custom cert/key pair is present.
func isManagedByMyController(cfg config.HttpsSSLConfig) bool {
	customCertFile := filepath.Join(cfg.CertDir, CustomCertFileName)
	customKeyFile := filepath.Join(cfg.CertDir, CustomKeyFileName)
	return !utils.IsFileExists(customCertFile) || !utils.IsFileExists(customKeyFile)
}

// GetSSLTLSConfig returns ssl certificate (startup-only path without daily renewal).
// Prefer NewSSLManager when the HTTPS listener needs hot-reload renewal.
func GetSSLTLSConfig(logger *zap.Logger, cfg config.HttpsSSLConfig) (*tls.Config, error) {
	m, err := NewSSLManager(logger, cfg, nil)
	if err != nil {
		return nil, err
	}
	return m.TLSConfig(), nil
}

// resolveValidityDays returns cfg value when positive, otherwise the default lifetime.
func resolveValidityDays(days int) int {
	if days > 0 {
		return days
	}
	return ValidityDays
}

// resolveRenewBeforeDays returns cfg value when positive, otherwise the default threshold.
func resolveRenewBeforeDays(days int) int {
	if days > 0 {
		return days
	}
	return RenewBeforeDays
}

// shouldRegenerateCert returns true when the generated cert/key are missing,
// cannot be parsed, or remaining validity is less than renewBeforeDays.
func shouldRegenerateCert(logger *zap.Logger, certFile, keyFile string, renewBeforeDays int) bool {
	if !utils.IsFileExists(certFile) || !utils.IsFileExists(keyFile) {
		logger.Info("generated SSL certificate not found, will create a new one", zap.String("certFile", certFile), zap.String("keyFile", keyFile))
		return true
	}

	remaining, err := certificateRemainingValidity(certFile)
	if err != nil {
		logger.Warn("unable to read existing SSL certificate, will create a new one", zap.String("certFile", certFile), zap.Error(err))
		return true
	}

	renewBefore := time.Duration(renewBeforeDays) * 24 * time.Hour
	if remaining < renewBefore {
		logger.Info("SSL certificate remaining validity is below threshold, will regenerate", zap.String("certFile", certFile),
			zap.Duration("remaining", remaining), zap.Int("renewBeforeDays", renewBeforeDays))
		return true
	}

	return false
}

// certificateRemainingValidity returns the time left until the certificate expires.
func certificateRemainingValidity(certFile string) (time.Duration, error) {
	certPEM, err := os.ReadFile(certFile)
	if err != nil {
		return 0, err
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return 0, errors.New("failed to decode PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return 0, err
	}

	return time.Until(cert.NotAfter), nil
}

// loadOrCreatePrivateKey reuses an existing RSA key when present and valid; otherwise creates one.
func loadOrCreatePrivateKey(logger *zap.Logger, certDir string) (*rsa.PrivateKey, bool, error) {
	keyPath := filepath.Join(certDir, GeneratedKeyFileName)
	if utils.IsFileExists(keyPath) {
		keyPEM, err := os.ReadFile(keyPath)
		if err == nil {
			block, _ := pem.Decode(keyPEM)
			if block != nil {
				if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
					if rsaKey, ok := key.(*rsa.PrivateKey); ok {
						logger.Info("reusing existing SSL private key", zap.String("keyFile", GeneratedKeyFileName))
						return rsaKey, true, nil
					}
				}
				// older OpenSSL-style PKCS#1 keys
				if rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
					logger.Info("reusing existing SSL private key", zap.String("keyFile", GeneratedKeyFileName))
					return rsaKey, true, nil
				}
			}
		}
		logger.Warn("unable to reuse existing SSL private key, will generate a new one", zap.String("keyFile", keyPath))
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, RSABits)
	if err != nil {
		return nil, false, err
	}
	return privateKey, false, nil
}

// generateSSLCert generates (or renews) the SSL certificate, reusing the private key when possible.
func generateSSLCert(logger *zap.Logger, certDir string, validityDays int) error {
	// reference: https://golang.org/src/crypto/tls/generate_cert.go
	privateKey, keyReused, err := loadOrCreatePrivateKey(logger, certDir)
	if err != nil {
		return err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		logger.Error("failed to generate serial number", zap.Error(err))
		return err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Hour * 24 * time.Duration(validityDays))

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{OrganizationName}},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		logger.Error("error on generating certificate", zap.Error(err))
		return err
	}

	var certBuf bytes.Buffer
	if err := pem.Encode(&certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}
	err = utils.WriteFile(certDir, GeneratedCertFileName, certBuf.Bytes())
	if err != nil {
		return err
	}

	if !keyReused {
		var keyBuf bytes.Buffer
		privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
		if err != nil {
			return err
		}
		if err := pem.Encode(&keyBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
			return err
		}

		err = utils.WriteFile(certDir, GeneratedKeyFileName, keyBuf.Bytes())
		if err != nil {
			return err
		}
	}

	logger.Info("generated new SSL certificate", zap.String("certDir", certDir), zap.String("certFile", GeneratedCertFileName), zap.String("keyFile", GeneratedKeyFileName),
		zap.Bool("keyReused", keyReused), zap.Time("notBefore", notBefore), zap.Time("notAfter", notAfter), zap.Int("validityDays", validityDays))

	return nil
}

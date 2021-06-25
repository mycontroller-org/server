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
	"time"

	"github.com/mycontroller-org/server/v2/pkg/model/config"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

const (
	CustomCertFileName = "custom.crt"
	CustomKeyFileName  = "custom.key"

	RSABits               = 2048
	OrganizationName      = "MyController.org"
	Validity              = 365 // days
	GeneratedCertFileName = "mc_generated.crt"
	GeneratedKeyFileName  = "mc_generated.key"
)

// GetSSLTLSConfig returns ssl certificate
func GetSSLTLSConfig(cfg config.HttpsSSLConfig) (*tls.Config, error) {

	if cfg.CertDir == "" {
		return nil, errors.New("cert_dir is missing")
	}

	certFile := fmt.Sprintf("%s/%s", cfg.CertDir, GeneratedCertFileName)
	keyFile := fmt.Sprintf("%s/%s", cfg.CertDir, GeneratedKeyFileName)

	// check the certificate on disk, if available use it and skip the following steps
	customCertFile := fmt.Sprintf("%s/%s", cfg.CertDir, CustomCertFileName)
	customKeyFile := fmt.Sprintf("%s/%s", cfg.CertDir, CustomKeyFileName)

	if utils.IsFileExists(customCertFile) && utils.IsFileExists(customKeyFile) {
		certFile = customCertFile
		keyFile = customKeyFile
	} else { // generate certificate
		err := generateSSLCert(cfg.CertDir)
		if err != nil {
			return nil, err
		}
	}

	tlsConfig := &tls.Config{Certificates: make([]tls.Certificate, 1)}
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig.Certificates[0] = certificate

	return tlsConfig, nil
}

// generateSSLCert generates ssl certificate and key
func generateSSLCert(certDir string) error {
	// check the certificate on disk, if available use it and skip the following steps

	// reference: https://golang.org/src/crypto/tls/generate_cert.go
	privateKey, err := rsa.GenerateKey(rand.Reader, RSABits)
	if err != nil {
		return err
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		zap.L().Error("failed to generate serial number", zap.Error(err))
		return err
	}

	template := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               pkix.Name{Organization: []string{OrganizationName}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24 * Validity),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		zap.L().Error("error on generating certificate", zap.Error(err))
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

	return nil
}

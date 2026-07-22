package https

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestShouldRegenerateCert_MissingFiles(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()
	certFile := filepath.Join(dir, GeneratedCertFileName)
	keyFile := filepath.Join(dir, GeneratedKeyFileName)

	if !shouldRegenerateCert(logger, certFile, keyFile, RenewBeforeDays) {
		t.Fatal("expected regeneration when cert and key files are missing")
	}
}

func TestShouldRegenerateCert_ValidCert(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()
	certFile, keyFile := writeTestCert(t, dir, time.Now(), time.Now().Add(365*24*time.Hour))

	if shouldRegenerateCert(logger, certFile, keyFile, RenewBeforeDays) {
		t.Fatal("expected existing long-lived cert to be reused")
	}
}

func TestShouldRegenerateCert_ExpiringSoon(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()
	// remaining validity: 10 days (< default RenewBeforeDays)
	certFile, keyFile := writeTestCert(t, dir, time.Now().Add(-355*24*time.Hour), time.Now().Add(10*24*time.Hour))

	if !shouldRegenerateCert(logger, certFile, keyFile, RenewBeforeDays) {
		t.Fatal("expected regeneration when remaining validity is less than renew_before_days")
	}
}

func TestShouldRegenerateCert_AboveThreshold(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()
	// remaining validity slightly above RenewBeforeDays so clock skew does not flip the result
	certFile, keyFile := writeTestCert(t, dir, time.Now().Add(-345*24*time.Hour), time.Now().Add(time.Duration(RenewBeforeDays+1)*24*time.Hour))

	if shouldRegenerateCert(logger, certFile, keyFile, RenewBeforeDays) {
		t.Fatal("expected cert to be reused when remaining validity is above the threshold")
	}
}

func TestShouldRegenerateCert_CustomRenewBeforeDays(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()
	// remaining validity: 15 days — regenerate only when threshold is higher than 15
	certFile, keyFile := writeTestCert(t, dir, time.Now().Add(-350*24*time.Hour), time.Now().Add(15*24*time.Hour))

	if shouldRegenerateCert(logger, certFile, keyFile, 10) {
		t.Fatal("expected cert to be reused when remaining validity is above custom threshold of 10 days")
	}
	if !shouldRegenerateCert(logger, certFile, keyFile, 30) {
		t.Fatal("expected regeneration when remaining validity is below custom threshold of 30 days")
	}
}

func TestShouldRegenerateCert_InvalidCert(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()
	certFile := filepath.Join(dir, GeneratedCertFileName)
	keyFile := filepath.Join(dir, GeneratedKeyFileName)

	if err := os.WriteFile(certFile, []byte("not-a-cert"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, []byte("not-a-key"), 0o600); err != nil {
		t.Fatal(err)
	}

	if !shouldRegenerateCert(logger, certFile, keyFile, RenewBeforeDays) {
		t.Fatal("expected regeneration when certificate cannot be parsed")
	}
}

func TestGetSSLTLSConfig_ReusesGeneratedCert(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()

	cfg := config.HttpsSSLConfig{CertDir: dir}
	tls1, err := GetSSLTLSConfig(logger, cfg)
	if err != nil {
		t.Fatalf("first GetSSLTLSConfig failed: %v", err)
	}
	if cert, err := tls1.GetCertificate(nil); err != nil || cert == nil {
		t.Fatalf("expected GetCertificate to return a cert: %v", err)
	}

	certPath := filepath.Join(dir, GeneratedCertFileName)
	keyPath := filepath.Join(dir, GeneratedKeyFileName)
	certBefore, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}
	keyBefore, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	// second call must reuse the same files
	tls2, err := GetSSLTLSConfig(logger, cfg)
	if err != nil {
		t.Fatalf("second GetSSLTLSConfig failed: %v", err)
	}
	if cert, err := tls2.GetCertificate(nil); err != nil || cert == nil {
		t.Fatalf("expected GetCertificate on reuse: %v", err)
	}

	certAfter, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}
	keyAfter, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(certBefore, certAfter) {
		t.Fatal("certificate was regenerated on restart; expected reuse")
	}
	if !bytes.Equal(keyBefore, keyAfter) {
		t.Fatal("private key was regenerated on restart; expected reuse")
	}
}

func TestGetSSLTLSConfig_RegeneratesWhenExpiring(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()

	writeTestCert(t, dir, time.Now().Add(-360*24*time.Hour), time.Now().Add(5*24*time.Hour))
	certPath := filepath.Join(dir, GeneratedCertFileName)
	keyPath := filepath.Join(dir, GeneratedKeyFileName)
	certBefore, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}
	keyBefore, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.HttpsSSLConfig{CertDir: dir}
	_, err = GetSSLTLSConfig(logger, cfg)
	if err != nil {
		t.Fatalf("GetSSLTLSConfig failed: %v", err)
	}

	certAfter, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}
	keyAfter, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(certBefore, certAfter) {
		t.Fatal("expected certificate to be regenerated when remaining validity is below threshold")
	}
	if !bytes.Equal(keyBefore, keyAfter) {
		t.Fatal("expected private key to be reused on renewal")
	}

	remaining, err := certificateRemainingValidity(certPath)
	if err != nil {
		t.Fatal(err)
	}
	minExpected := time.Duration(ValidityDays-1) * 24 * time.Hour
	if remaining < minExpected {
		t.Fatalf("regenerated cert remaining validity too short: %v", remaining)
	}
}

func TestGetSSLTLSConfig_CustomValidityAndRenewBefore(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()

	writeTestCert(t, dir, time.Now().Add(-340*24*time.Hour), time.Now().Add(25*24*time.Hour))
	certPath := filepath.Join(dir, GeneratedCertFileName)
	certBefore, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}

	cfg := config.HttpsSSLConfig{
		CertDir:         dir,
		ValidityDays:    90,
		RenewBeforeDays: 30,
	}
	_, err = GetSSLTLSConfig(logger, cfg)
	if err != nil {
		t.Fatalf("GetSSLTLSConfig failed: %v", err)
	}

	certAfter, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(certBefore, certAfter) {
		t.Fatal("expected regeneration with custom renew_before_days=30")
	}

	remaining, err := certificateRemainingValidity(certPath)
	if err != nil {
		t.Fatal(err)
	}
	minExpected := 89 * 24 * time.Hour
	maxExpected := 90*24*time.Hour + time.Hour
	if remaining < minExpected || remaining > maxExpected {
		t.Fatalf("expected ~90 days validity from config, got %v", remaining)
	}
}

func TestGetSSLTLSConfig_DefaultsWhenUnset(t *testing.T) {
	if got := resolveValidityDays(0); got != ValidityDays {
		t.Fatalf("expected default validity %d, got %d", ValidityDays, got)
	}
	if got := resolveValidityDays(-1); got != ValidityDays {
		t.Fatalf("expected default validity for negative, got %d", got)
	}
	if got := resolveValidityDays(180); got != 180 {
		t.Fatalf("expected custom validity 180, got %d", got)
	}
	if got := resolveRenewBeforeDays(0); got != RenewBeforeDays {
		t.Fatalf("expected default renew_before %d, got %d", RenewBeforeDays, got)
	}
	if got := resolveRenewBeforeDays(7); got != 7 {
		t.Fatalf("expected custom renew_before 7, got %d", got)
	}
}

func TestNewSSLManager_WarnsOnInvalidManagedConfig(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	logger := zap.New(core)
	dir := t.TempDir()

	// equal
	_, err := NewSSLManager(logger, config.HttpsSSLConfig{
		CertDir:         dir,
		ValidityDays:    30,
		RenewBeforeDays: 30,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// renew_before > validity
	_, err = NewSSLManager(logger, config.HttpsSSLConfig{
		CertDir:         t.TempDir(),
		ValidityDays:    15,
		RenewBeforeDays: 30,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// valid config should not warn
	_, err = NewSSLManager(logger, config.HttpsSSLConfig{
		CertDir:         t.TempDir(),
		ValidityDays:    365,
		RenewBeforeDays: 30,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	warnCount := 0
	for _, e := range logs.All() {
		if e.Level == zap.WarnLevel {
			warnCount++
		}
	}
	if warnCount < 2 {
		t.Fatalf("expected at least 2 warnings for invalid configs, got %d", warnCount)
	}
}

func TestGenerateSSLCert_ValidityOneYear(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()

	if err := generateSSLCert(logger, dir, ValidityDays); err != nil {
		t.Fatalf("generateSSLCert failed: %v", err)
	}

	certPath := filepath.Join(dir, GeneratedCertFileName)
	remaining, err := certificateRemainingValidity(certPath)
	if err != nil {
		t.Fatal(err)
	}

	minExpected := time.Duration(ValidityDays-1) * 24 * time.Hour
	maxExpected := time.Duration(ValidityDays)*24*time.Hour + time.Hour
	if remaining < minExpected || remaining > maxExpected {
		t.Fatalf("expected ~%d days validity, got %v", ValidityDays, remaining)
	}
}

func TestGenerateSSLCert_CustomValidity(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()
	const days = 30

	if err := generateSSLCert(logger, dir, days); err != nil {
		t.Fatalf("generateSSLCert failed: %v", err)
	}

	certPath := filepath.Join(dir, GeneratedCertFileName)
	remaining, err := certificateRemainingValidity(certPath)
	if err != nil {
		t.Fatal(err)
	}

	minExpected := time.Duration(days-1) * 24 * time.Hour
	maxExpected := time.Duration(days)*24*time.Hour + time.Hour
	if remaining < minExpected || remaining > maxExpected {
		t.Fatalf("expected ~%d days validity, got %v", days, remaining)
	}
}

func TestGenerateSSLCert_ReusesPrivateKey(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()

	if err := generateSSLCert(logger, dir, ValidityDays); err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(dir, GeneratedKeyFileName)
	keyBefore, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := generateSSLCert(logger, dir, ValidityDays); err != nil {
		t.Fatal(err)
	}
	keyAfter, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(keyBefore, keyAfter) {
		t.Fatal("expected private key file to be unchanged on renew")
	}
}

func TestSSLManager_ManagedVsCustom(t *testing.T) {
	logger := zap.NewNop()

	// managed: no custom files
	managedDir := t.TempDir()
	m, err := NewSSLManager(logger, config.HttpsSSLConfig{CertDir: managedDir}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !m.Managed() {
		t.Fatal("expected managed certificate when custom files are absent")
	}

	// custom: both custom.crt and custom.key present
	customDir := t.TempDir()
	writeCustomCert(t, customDir)
	c, err := NewSSLManager(logger, config.HttpsSSLConfig{CertDir: customDir}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Managed() {
		t.Fatal("expected custom certificate when custom.crt and custom.key exist")
	}
}

func TestSSLManager_CheckAndRenew_HotReload(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()

	// seed an expiring managed cert
	writeTestCert(t, dir, time.Now().Add(-360*24*time.Hour), time.Now().Add(5*24*time.Hour))
	certPath := filepath.Join(dir, GeneratedCertFileName)
	keyPath := filepath.Join(dir, GeneratedKeyFileName)
	keyBefore, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	m, err := NewSSLManager(logger, config.HttpsSSLConfig{CertDir: dir}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !m.Managed() {
		t.Fatal("expected managed SSL")
	}

	// force another near-expiry and call CheckAndRenew again
	writeTestCert(t, dir, time.Now().Add(-360*24*time.Hour), time.Now().Add(5*24*time.Hour))
	// restore original key so renew reuses it (writeTestCert overwrites key)
	if err := os.WriteFile(keyPath, keyBefore, 0o600); err != nil {
		t.Fatal(err)
	}

	beforeReload, err := m.TLSConfig().GetCertificate(nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := m.CheckAndRenew(); err != nil {
		t.Fatalf("CheckAndRenew failed: %v", err)
	}

	afterReload, err := m.TLSConfig().GetCertificate(nil)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(beforeReload.Certificate[0], afterReload.Certificate[0]) {
		t.Fatal("expected in-memory certificate to be hot-reloaded after CheckAndRenew")
	}

	remaining, err := certificateRemainingValidity(certPath)
	if err != nil {
		t.Fatal(err)
	}
	if remaining < time.Duration(ValidityDays-1)*24*time.Hour {
		t.Fatalf("expected renewed long-lived cert, remaining=%v", remaining)
	}

	keyAfter, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(keyBefore, keyAfter) {
		t.Fatal("expected key reuse on CheckAndRenew")
	}
}

func TestSSLManager_CheckAndRenew_SkipsWhenValid(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()
	cfg := config.HttpsSSLConfig{CertDir: dir}

	m, err := NewSSLManager(logger, cfg, nil)
	if err != nil {
		t.Fatal(err)
	}

	certPath := filepath.Join(dir, GeneratedCertFileName)
	certBefore, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := m.CheckAndRenew(); err != nil {
		t.Fatal(err)
	}

	certAfter, err := os.ReadFile(certPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(certBefore, certAfter) {
		t.Fatal("did not expect renewal when certificate is still valid")
	}
}

func TestSSLManager_StartDailyRenewalCheck_Managed(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()
	sched := &fakeScheduler{}

	m, err := NewSSLManager(logger, config.HttpsSSLConfig{CertDir: dir}, sched)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.StartDailyRenewalCheck(); err != nil {
		t.Fatal(err)
	}
	if !sched.has(sslRenewalJobName) {
		t.Fatal("expected daily renewal job to be scheduled for managed SSL")
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	if sched.has(sslRenewalJobName) {
		t.Fatal("expected daily renewal job to be removed on Close")
	}
}

func TestSSLManager_StartDailyRenewalCheck_CustomDisabled(t *testing.T) {
	logger := zap.NewNop()
	dir := t.TempDir()
	writeCustomCert(t, dir)
	sched := &fakeScheduler{}

	m, err := NewSSLManager(logger, config.HttpsSSLConfig{CertDir: dir}, sched)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.StartDailyRenewalCheck(); err != nil {
		t.Fatal(err)
	}
	if sched.has(sslRenewalJobName) {
		t.Fatal("expected no daily renewal job for custom certificates")
	}
}

// fakeScheduler records scheduled jobs for unit tests.
type fakeScheduler struct {
	mu   sync.Mutex
	jobs map[string]string
}

func (f *fakeScheduler) Name() string { return "fake" }
func (f *fakeScheduler) Start() error { return nil }
func (f *fakeScheduler) Close() error { return nil }
func (f *fakeScheduler) ListNames() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	names := make([]string, 0, len(f.jobs))
	for n := range f.jobs {
		names = append(names, n)
	}
	return names
}
func (f *fakeScheduler) IsAvailable(id string) bool { return f.has(id) }
func (f *fakeScheduler) RemoveWithPrefix(prefix string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for n := range f.jobs {
		if len(n) >= len(prefix) && n[:len(prefix)] == prefix {
			delete(f.jobs, n)
		}
	}
}
func (f *fakeScheduler) AddFunc(name, spec string, _ func()) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.jobs == nil {
		f.jobs = map[string]string{}
	}
	f.jobs[name] = spec
	return nil
}
func (f *fakeScheduler) RemoveFunc(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.jobs, name)
}
func (f *fakeScheduler) has(name string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.jobs[name]
	return ok
}

// writeCustomCert writes custom.crt / custom.key under dir.
func writeCustomCert(t *testing.T, dir string) {
	t.Helper()
	certFile, keyFile := writeTestCert(t, dir, time.Now(), time.Now().Add(365*24*time.Hour))
	// move generated names to custom names
	customCert := filepath.Join(dir, CustomCertFileName)
	customKey := filepath.Join(dir, CustomKeyFileName)
	if err := os.Rename(certFile, customCert); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(keyFile, customKey); err != nil {
		t.Fatal(err)
	}
}

// writeTestCert creates a self-signed cert/key pair under dir with the given validity window.
func writeTestCert(t *testing.T, dir string, notBefore, notAfter time.Time) (certFile, keyFile string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		t.Fatal(err)
	}

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
		t.Fatal(err)
	}

	var certBuf bytes.Buffer
	if err := pem.Encode(&certBuf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		t.Fatal(err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	var keyBuf bytes.Buffer
	if err := pem.Encode(&keyBuf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		t.Fatal(err)
	}

	certFile = filepath.Join(dir, GeneratedCertFileName)
	keyFile = filepath.Join(dir, GeneratedKeyFileName)
	if err := os.WriteFile(certFile, certBuf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, keyBuf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	return certFile, keyFile
}

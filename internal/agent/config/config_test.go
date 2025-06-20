package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/sender"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	t.Setenv("ADDRESS", "test.host:1234")
	t.Setenv("POLL_INTERVAL", "3")
	t.Setenv("REPORT_INTERVAL", "11")
	t.Setenv("KEY", "test_KEY")
	t.Setenv("RATE_LIMIT", "15")
	t.Setenv("BATCHING", "true")
	t.Setenv("PPROF_ADDRESS", ":6066")
	t.Setenv("CRYPTO_KEY", "")
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"test"}
	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
	cfg, err := NewConfig()
	require.NoError(t, err)
	assert.Equal(t, "test.host:1234", cfg.Addr)
	assert.Equal(t, 3, cfg.PollInterval)
	assert.Equal(t, 11, cfg.ReportInterval)
	assert.Equal(t, "test_KEY", cfg.Key)
	assert.Equal(t, 15, cfg.RateLimit)
	assert.True(t, cfg.Batching)
	assert.Equal(t, ":6066", cfg.PprofAddr)
	assert.Equal(t, sender.Config{
		UpdateURL:   "http://" + cfg.Addr + "/update/",
		UpdatesURL:  "http://" + cfg.Addr + "/updates/",
		RetryDelays: []time.Duration{time.Second, 3 * time.Second, 5 * time.Second},
		Timeout:     10 * time.Second,
		Key:         []byte(cfg.Key),
		RateLimit:   cfg.RateLimit,
	}, *cfg.Sender)
}

func TestNewConfig_CryptoKey(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"test"}
	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
	tempDir, err := os.MkdirTemp("", "rsa_test")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tempDir))
	}()
	privatePath := filepath.Join(tempDir, "private_key.pem")
	publicPath := filepath.Join(tempDir, "public_key.pem")
	require.NoError(t, generateTempRSA(t, 2048, privatePath, publicPath))

	t.Setenv("CRYPTO_KEY", publicPath)
	cfg, err := NewConfig()
	require.NoError(t, err)
	assert.Equal(t, publicPath, cfg.CryptoKey)
	assert.NotNil(t, cfg.Sender.PublicKey)
}

func TestNewConfig_CryptoKey_not_exists(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"test"}
	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
	t.Setenv("CRYPTO_KEY", "not_exist.pem")
	cfg, err := NewConfig()
	assert.Error(t, err)
	assert.Equal(t, "not_exist.pem", cfg.CryptoKey)
	assert.Nil(t, cfg.Sender.PublicKey)
}

func TestNewConfig_CryptoKey_with_private_key(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"test"}
	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
	tempDir, err := os.MkdirTemp("", "rsa_test")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tempDir))
	}()
	privatePath := filepath.Join(tempDir, "private_key.pem")
	publicPath := filepath.Join(tempDir, "public_key.pem")
	require.NoError(t, generateTempRSA(t, 2048, privatePath, publicPath))
	t.Setenv("CRYPTO_KEY", privatePath)
	cfg, err := NewConfig()
	assert.Error(t, err)
	assert.Equal(t, privatePath, cfg.CryptoKey)
	assert.Nil(t, cfg.Sender.PublicKey)
}

func TestNewConfig_CryptoKey_invalid_key(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"test"}
	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
	tempDir, err := os.MkdirTemp("", "rsa_test")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tempDir))
	}()
	privatePath := filepath.Join(tempDir, "private_key.pem")
	publicPath := filepath.Join(tempDir, "public_key.pem")
	require.NoError(t, generateTempRSA(t, 2048, privatePath, publicPath))
	publicKeyData, err := os.ReadFile(publicPath)
	require.NoError(t, err)
	publicKeyData[34] = '*'
	require.NoError(t, os.WriteFile(publicPath, publicKeyData, 0600))
	t.Setenv("CRYPTO_KEY", publicPath)
	cfg, err := NewConfig()
	assert.Error(t, err)
	assert.Equal(t, publicPath, cfg.CryptoKey)
	assert.Nil(t, cfg.Sender.PublicKey)
}

func generateTempRSA(t *testing.T, bits int, privatePath, publicPath string) error {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	require.NoError(t, err)
	publicKey := &privateKey.PublicKey
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes}
	privateFile, err := os.Create(privatePath)
	require.NoError(t, err)
	defer func() {
		_ = privateFile.Close()
	}()
	require.NoError(t, pem.Encode(privateFile, privateKeyBlock))
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	require.NoError(t, err)
	publicKeyBlock := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: publicKeyBytes}
	publicFile, err := os.Create(publicPath)
	require.NoError(t, err)
	defer func() {
		_ = publicFile.Close()
	}()
	require.NoError(t, pem.Encode(publicFile, publicKeyBlock))
	return nil
}

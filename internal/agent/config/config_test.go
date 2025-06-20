package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
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
		Timeout:     11 * time.Second,
		Key:         []byte(cfg.Key),
		RateLimit:   cfg.RateLimit,
	}, *cfg.Sender)
}

func TestNewConfig_JSON(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	tempDir, err := os.MkdirTemp("", "config_test")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() { assert.NoError(t, os.RemoveAll(tempDir)) }()
	configPath := filepath.Join(tempDir, "config.json")

	jsonConfig := map[string]interface{}{
		"address":         "json:8080",
		"poll_interval":   "3s",
		"report_interval": "9s",
		"crypto_key":      "",
	}
	jsonData, err := json.Marshal(jsonConfig)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, jsonData, 0600))

	tests := []struct {
		envVars        map[string]string
		wantCfg        *Config
		name           string
		configPath     string
		wantErrMessage string
		args           []string
		wantErr        bool
	}{
		{
			name:       "valid JSON config",
			configPath: configPath,
			args:       []string{"test", "-config", configPath},
			wantCfg: &Config{
				Addr:           "json:8080",
				PollInterval:   3,
				ReportInterval: 9,
			},
			wantErr: false,
		},
		{
			name:       "JSON with env override",
			configPath: configPath,
			args:       []string{"test", "-config", configPath},
			envVars: map[string]string{
				"ADDRESS":         "env:8080",
				"POLL_INTERVAL":   "5",
				"REPORT_INTERVAL": "15",
			},
			wantCfg: &Config{
				Addr:           "env:8080",
				PollInterval:   5,
				ReportInterval: 15,
			},
			wantErr: false,
		},
		{
			name:           "non-existent JSON file",
			configPath:     filepath.Join(tempDir, "nonexistent.json"),
			args:           []string{"test", "-config", filepath.Join(tempDir, "nonexistent.json")},
			wantErr:        true,
			wantErrMessage: "failed to stat file",
		},
		{
			name:           "invalid JSON",
			configPath:     filepath.Join(tempDir, "invalid.json"),
			args:           []string{"test", "-config", filepath.Join(tempDir, "invalid.json")},
			wantErr:        true,
			wantErrMessage: "failed to unmarshal JSON",
		},
		{
			name:           "invalid PollIntervalStr",
			configPath:     filepath.Join(tempDir, "invalid_interval.json"),
			args:           []string{"test", "-config", filepath.Join(tempDir, "invalid_interval.json")},
			wantErr:        true,
			wantErrMessage: "failed to parse poll duration",
		},
		{
			name:           "invalid ReportIntervalStr",
			configPath:     filepath.Join(tempDir, "invalid_interval.json"),
			args:           []string{"test", "-config", filepath.Join(tempDir, "invalid_interval.json")},
			wantErr:        true,
			wantErrMessage: "failed to parse report duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
			if tt.name == "invalid JSON" {
				require.NoError(t, os.WriteFile(tt.configPath, []byte("{invalid}"), 0600), "Failed to write invalid JSON")
			} else if tt.name == "invalid PollIntervalStr" {
				invalidJSON := map[string]interface{}{
					"poll_interval": "invalid",
				}
				data, err := json.Marshal(invalidJSON)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(tt.configPath, data, 0600), "Failed to write invalid interval JSON")
			} else if tt.name == "invalid ReportIntervalStr" {
				invalidJSON := map[string]interface{}{
					"report_interval": "invalid",
				}
				data, err := json.Marshal(invalidJSON)
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(tt.configPath, data, 0600), "Failed to write invalid interval JSON")
			}

			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}
			os.Args = tt.args

			cfg, err := NewConfig()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCfg.Addr, cfg.Addr)
				assert.Equal(t, tt.wantCfg.PollInterval, cfg.PollInterval)
				assert.Equal(t, tt.wantCfg.ReportInterval, cfg.ReportInterval)
			}
		})
	}
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

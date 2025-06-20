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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	t.Setenv("ADDRESS", "test_ADDRESS")
	t.Setenv("FILE_STORAGE_PATH", "test_FILE_STORAGE_PATH")
	t.Setenv("DATABASE_DSN", "test_DATABASE_DSN")
	t.Setenv("RESTORE", "true")
	t.Setenv("STORE_INTERVAL", "5")
	t.Setenv("KEY", "test_KEY")
	t.Setenv("PPROF", "true")
	t.Setenv("CRYPTO_KEY", "")
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"test"}
	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
	cfg, err := NewConfig()
	require.NoError(t, err)
	assert.Equal(t, "test_ADDRESS", cfg.Addr)
	assert.Equal(t, "test_FILE_STORAGE_PATH", cfg.FileStoragePath)
	assert.Equal(t, "test_DATABASE_DSN", cfg.DatabaseDSN)
	assert.True(t, cfg.Restore)
	assert.Equal(t, int64(5), cfg.StoreInterval)
	assert.Equal(t, "test_KEY", cfg.Key)
	assert.Equal(t, "", cfg.CryptoKey)
	assert.True(t, cfg.Pprof)
	assert.Equal(t, 5*time.Second, cfg.ShutdownTimeout)
	assert.Equal(t, 5*time.Second, cfg.DatabasePingTimeout)
	assert.Equal(t, []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}, cfg.RetryDelays)
}

func TestNewConfig_JSON(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	tempDir, err := os.MkdirTemp("", "config_test")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() { assert.NoError(t, os.RemoveAll(tempDir)) }()
	configPath := filepath.Join(tempDir, "config.json")

	jsonConfig := map[string]interface{}{
		"address":        "json:8080",
		"store_file":     "json_storage.json",
		"database_dsn":   "json_dsn",
		"restore":        false,
		"store_interval": "10s",
		"crypto_key":     "",
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
				Addr:            "json:8080",
				FileStoragePath: "json_storage.json",
				DatabaseDSN:     "json_dsn",
				Restore:         false,
				StoreInterval:   10,
			},
			wantErr: false,
		},
		{
			name:       "JSON with env override",
			configPath: configPath,
			args:       []string{"test", "-config", configPath},
			envVars: map[string]string{
				"ADDRESS":        "env:8080",
				"STORE_INTERVAL": "5",
			},
			wantCfg: &Config{
				Addr:            "env:8080",
				FileStoragePath: "json_storage.json",
				DatabaseDSN:     "json_dsn",
				Restore:         false,
				StoreInterval:   5,
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
			name:           "invalid StoreIntervalStr",
			configPath:     filepath.Join(tempDir, "invalid_interval.json"),
			args:           []string{"test", "-config", filepath.Join(tempDir, "invalid_interval.json")},
			wantErr:        true,
			wantErrMessage: "failed to parse duration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
			if tt.name == "invalid JSON" {
				require.NoError(t, os.WriteFile(tt.configPath, []byte("{invalid}"), 0600), "Failed to write invalid JSON")
			} else if tt.name == "invalid StoreIntervalStr" {
				invalidJSON := map[string]interface{}{
					"store_interval": "invalid",
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
				assert.Equal(t, tt.wantCfg.FileStoragePath, cfg.FileStoragePath)
				assert.Equal(t, tt.wantCfg.DatabaseDSN, cfg.DatabaseDSN)
				assert.Equal(t, tt.wantCfg.Restore, cfg.Restore)
				assert.Equal(t, tt.wantCfg.StoreInterval, cfg.StoreInterval)
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

	t.Setenv("CRYPTO_KEY", privatePath)
	cfg, err := NewConfig()
	require.NoError(t, err)
	assert.Equal(t, privatePath, cfg.CryptoKey)
	assert.NotNil(t, cfg.PrivateKey)
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
	assert.Nil(t, cfg.PrivateKey)
}

func TestNewConfig_CryptoKey_with_public_key(t *testing.T) {
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
	assert.Error(t, err)
	assert.Equal(t, publicPath, cfg.CryptoKey)
	assert.Nil(t, cfg.PrivateKey)
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
	privateKeyData, err := os.ReadFile(privatePath)
	require.NoError(t, err)
	privateKeyData[34] = '*'
	require.NoError(t, os.WriteFile(privatePath, privateKeyData, 0600))
	t.Setenv("CRYPTO_KEY", privatePath)
	cfg, err := NewConfig()
	assert.Error(t, err)
	assert.Equal(t, privatePath, cfg.CryptoKey)
	assert.Nil(t, cfg.PrivateKey)
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

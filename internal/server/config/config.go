// Package config contains the config logic.
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

// Config is the server config.
type Config struct {
	PrivateKey          *rsa.PrivateKey
	Addr                string `env:"ADDRESS"`
	FileStoragePath     string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN         string `env:"DATABASE_DSN"`
	Key                 string `env:"KEY"`
	CryptoKey           string `env:"CRYPTO_KEY"`
	RetryDelays         []time.Duration
	StoreInterval       int64 `env:"STORE_INTERVAL"`
	ShutdownTimeout     time.Duration
	DatabasePingTimeout time.Duration
	Restore             bool `env:"RESTORE"`
	Pprof               bool `env:"PPROF"`
}

// NewConfig returns the server config.
func NewConfig() (*Config, error) {
	const (
		storeInterval   = 0
		shutdownTimeout = 5
		databasePingTimeout
	)
	cfg := &Config{}
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server host")
	flag.StringVar(&cfg.FileStoragePath, "f", "storage.json", "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn")
	flag.BoolVar(&cfg.Restore, "r", true, "file storage path")
	flag.Int64Var(&cfg.StoreInterval, "i", storeInterval, "store interval")
	flag.StringVar(&cfg.Key, "k", "", "key")
	flag.BoolVar(&cfg.Pprof, "pprof", false, "use pprof")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "crypto key")

	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.ShutdownTimeout = shutdownTimeout * time.Second
	cfg.DatabasePingTimeout = databasePingTimeout * time.Second
	cfg.RetryDelays = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

	if err = cfg.loadPrivateKey(); err != nil {
		return cfg, fmt.Errorf("failed to load private key: %w", err)
	}

	return cfg, nil
}

func (cfg *Config) loadPrivateKey() error {
	if cfg.CryptoKey == "" {
		return nil
	}
	var err error
	if _, err = os.Stat(cfg.CryptoKey); os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist: %w", cfg.CryptoKey, err)
	} else if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", cfg.CryptoKey, err)
	}
	data, err := os.ReadFile(cfg.CryptoKey)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", cfg.CryptoKey, err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block from file %s", cfg.CryptoKey)
	}
	if block.Type != "RSA PRIVATE KEY" {
		return fmt.Errorf("file %s is not a RSA private key", cfg.CryptoKey)
	}
	cfg.PrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("file %s is not a valid RSA private key: %w", cfg.CryptoKey, err)
	}
	return nil
}

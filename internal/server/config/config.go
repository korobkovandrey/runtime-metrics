// Package config contains the config logic.
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
)

// Config is the server config.
type Config struct {
	PrivateKey          *rsa.PrivateKey
	StoreIntervalStr    *string `json:"store_interval"`
	IPNet               *net.IPNet
	TrustedSubnet       string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	Addr                string `env:"ADDRESS" json:"address"`
	GRPSAddr            string `env:"GRPC_ADDRESS" json:"grpc_address"`
	FileStoragePath     string `env:"FILE_STORAGE_PATH" json:"store_file"`
	DatabaseDSN         string `env:"DATABASE_DSN" json:"database_dsn"`
	Key                 string `env:"KEY"`
	CryptoKey           string `env:"CRYPTO_KEY" json:"crypto_key"`
	ConfigPath          string `env:"CONFIG"`
	RetryDelays         []time.Duration
	StoreInterval       int64 `env:"STORE_INTERVAL"`
	DatabasePingTimeout time.Duration
	ShutdownTimeout     time.Duration
	Restore             bool `env:"RESTORE" json:"restore"`
	Pprof               bool `env:"PPROF"`
}

// NewConfig returns the server config.
func NewConfig() (*Config, error) {
	const (
		shutdownTimeout = 5
		databasePingTimeout
	)
	cfg := &Config{
		Addr:            "localhost:8080",
		GRPSAddr:        "localhost:3200",
		FileStoragePath: "storage.json",
		Restore:         true,
	}
	err := loadJSONConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load json config: %w", err)
	}
	err = parseFlags(cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to parse flags: %w", err)
	}
	cfg.ShutdownTimeout = shutdownTimeout * time.Second
	cfg.DatabasePingTimeout = databasePingTimeout * time.Second
	cfg.RetryDelays = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

	if cfg.TrustedSubnet != "" {
		_, cfg.IPNet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse trusted subnet: %w", err)
		}
	}

	if err = cfg.loadPrivateKey(); err != nil {
		return cfg, fmt.Errorf("failed to load private key: %w", err)
	}
	return cfg, nil
}

func parseFlags(cfg *Config) error {
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "HTTP server host")
	flag.StringVar(&cfg.GRPSAddr, "g", cfg.GRPSAddr, "GRPS server host")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database dsn")
	flag.BoolVar(&cfg.Restore, "r", cfg.Restore, "file storage path")
	flag.Int64Var(&cfg.StoreInterval, "i", cfg.StoreInterval, "store interval")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "key")
	flag.BoolVar(&cfg.Pprof, "pprof", cfg.Pprof, "use pprof")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "crypto key")
	flag.StringVar(&cfg.ConfigPath, "config", "", "config path")
	flag.StringVar(&cfg.ConfigPath, "c", "", "config path")
	flag.StringVar(&cfg.TrustedSubnet, "t", "", "trusted subnet")
	flag.Parse()
	err := env.Parse(cfg)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	return nil
}

func loadJSONConfig(cfg *Config) error {
	cfg.ConfigPath = os.Getenv("CONFIG")
	if cfg.ConfigPath == "" {
		if err := parseFlags(cfg); err != nil {
			return fmt.Errorf("failed to parse flags: %w", err)
		}
		flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
		if cfg.ConfigPath == "" {
			return nil
		}
	}
	if _, err := os.Stat(cfg.ConfigPath); err != nil {
		return fmt.Errorf("failed to stat file %s: %w", cfg.ConfigPath, err)
	}
	jsonDataByte, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", cfg.ConfigPath, err)
	}
	if err = json.Unmarshal(jsonDataByte, &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	if cfg.StoreIntervalStr != nil {
		var d time.Duration
		d, err = time.ParseDuration(*cfg.StoreIntervalStr)
		if err != nil {
			return fmt.Errorf("failed to parse duration: %w", err)
		}
		cfg.StoreInterval = int64(d.Seconds())
	}
	return nil
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

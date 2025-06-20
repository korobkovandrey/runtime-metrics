// Package config contains the config logic.
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/sender"
)

// Config is the agent config.
type Config struct {
	Sender         *sender.Config
	Addr           string `env:"ADDRESS"`
	Key            string `env:"KEY"`
	PprofAddr      string `env:"PPROF_ADDRESS"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	RateLimit      int    `env:"RATE_LIMIT"`
	Batching       bool   `env:"BATCHING"`
}

// NewConfig returns the agent config.
func NewConfig() (*Config, error) {
	const (
		pollIntervalSeconds   = 2
		reportIntervalSeconds = 10
	)
	cfg := &Config{}
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server host")
	flag.IntVar(&cfg.PollInterval, "p", pollIntervalSeconds, "pollInterval in seconds")
	flag.IntVar(&cfg.ReportInterval, "r", reportIntervalSeconds, "reportInterval in seconds")
	flag.StringVar(&cfg.Key, "k", "", "key")
	flag.IntVar(&cfg.RateLimit, "l", runtime.NumCPU(), "rate limit")
	flag.BoolVar(&cfg.Batching, "b", true, "batching")
	flag.StringVar(&cfg.PprofAddr, "pprof", "", "pprof address")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "crypto key")

	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.ReportInterval < 1 {
		return cfg, fmt.Errorf("ReportInterval (%ds) must be greater 0",
			cfg.ReportInterval)
	}

	if cfg.PollInterval < 1 {
		return cfg, fmt.Errorf("ReportInterval (%ds) must be greater 0",
			cfg.ReportInterval)
	}

	if cfg.ReportInterval <= cfg.PollInterval {
		return cfg, fmt.Errorf("ReportInterval (%ds) must be greater than PollInterval (%ds)",
			cfg.ReportInterval, cfg.PollInterval)
	}

	if cfg.RateLimit < 1 {
		return cfg, fmt.Errorf("RateLimit (%d) must be greater 0",
			cfg.RateLimit)
	}

	baseURL := "http://" + cfg.Addr
	cfg.Sender = &sender.Config{
		UpdateURL:   baseURL + "/update/",
		UpdatesURL:  baseURL + "/updates/",
		RetryDelays: []time.Duration{time.Second, 3 * time.Second, 5 * time.Second},
		Timeout:     reportIntervalSeconds * time.Second,
		Key:         []byte(cfg.Key),
		RateLimit:   cfg.RateLimit,
	}
	if err = cfg.loadPublicKey(); err != nil {
		return cfg, fmt.Errorf("failed to load public key: %w", err)
	}
	return cfg, nil
}

func (cfg *Config) loadPublicKey() error {
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
	if block.Type != "RSA PUBLIC KEY" {
		return fmt.Errorf("file %s is not a RSA public key", cfg.CryptoKey)
	}
	parsedPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("file %s is not a valid RSA public key: %w", cfg.CryptoKey, err)
	}
	var ok bool
	cfg.Sender.PublicKey, ok = parsedPublicKey.(*rsa.PublicKey)
	if !ok {
		cfg.Sender.PublicKey = nil
		return fmt.Errorf("file %s is not a valid RSA public key", cfg.CryptoKey)
	}
	return nil
}

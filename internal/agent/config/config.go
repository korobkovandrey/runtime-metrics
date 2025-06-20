// Package config contains the config logic.
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
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
	Sender            *sender.Config
	PollIntervalStr   *string `json:"poll_interval"`
	ReportIntervalStr *string `json:"report_interval"`
	Addr              string  `env:"ADDRESS" json:"address"`
	Key               string  `env:"KEY"`
	PprofAddr         string  `env:"PPROF_ADDRESS"`
	CryptoKey         string  `env:"CRYPTO_KEY" json:"crypto_key"`
	ConfigPath        string  `env:"CONFIG"`
	PollInterval      int     `env:"POLL_INTERVAL"`
	ReportInterval    int     `env:"REPORT_INTERVAL"`
	RateLimit         int     `env:"RATE_LIMIT"`
	Batching          bool    `env:"BATCHING"`
}

// NewConfig returns the agent config.
func NewConfig() (*Config, error) {
	const (
		pollIntervalSeconds   = 2
		reportIntervalSeconds = 10
	)
	cfg := &Config{
		Addr:           "localhost:8080",
		PollInterval:   pollIntervalSeconds,
		ReportInterval: reportIntervalSeconds,
	}
	err := loadJSONConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load json config: %w", err)
	}
	err = parseFlags(cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to parse flags: %w", err)
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
		Timeout:     time.Duration(cfg.ReportInterval) * time.Second,
		Key:         []byte(cfg.Key),
		RateLimit:   cfg.RateLimit,
	}
	if err = cfg.loadPublicKey(); err != nil {
		return cfg, fmt.Errorf("failed to load public key: %w", err)
	}
	return cfg, nil
}

func parseFlags(cfg *Config) error {
	flag.StringVar(&cfg.Addr, "a", cfg.Addr, "server host")
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "pollInterval in seconds")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "reportInterval in seconds")
	flag.StringVar(&cfg.Key, "k", "", "key")
	flag.IntVar(&cfg.RateLimit, "l", runtime.NumCPU(), "rate limit")
	flag.BoolVar(&cfg.Batching, "b", true, "batching")
	flag.StringVar(&cfg.PprofAddr, "pprof", "", "pprof address")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "crypto key")
	flag.StringVar(&cfg.ConfigPath, "config", "", "config path")
	flag.StringVar(&cfg.ConfigPath, "c", "", "config path")
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
	if cfg.PollIntervalStr != nil {
		var d time.Duration
		d, err = time.ParseDuration(*cfg.PollIntervalStr)
		if err != nil {
			return fmt.Errorf("failed to parse poll duration: %w", err)
		}
		cfg.PollInterval = int(d.Seconds())
	}
	if cfg.ReportIntervalStr != nil {
		var d time.Duration
		d, err = time.ParseDuration(*cfg.ReportIntervalStr)
		if err != nil {
			return fmt.Errorf("failed to parse report duration: %w", err)
		}
		cfg.ReportInterval = int(d.Seconds())
	}
	return nil
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

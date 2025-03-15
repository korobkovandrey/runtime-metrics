package config

import (
	"flag"
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr                string `env:"ADDRESS"`
	FileStoragePath     string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN         string `env:"DATABASE_DSN"`
	Restore             bool   `env:"RESTORE"`
	StoreInterval       int64  `env:"STORE_INTERVAL"`
	ShutdownTimeout     time.Duration
	DatabasePingTimeout time.Duration
	RetryDelays         []time.Duration
}

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

	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return cfg, fmt.Errorf("NewConfig: %w", err)
	}

	cfg.ShutdownTimeout = shutdownTimeout * time.Second
	cfg.DatabasePingTimeout = databasePingTimeout * time.Second
	cfg.RetryDelays = []time.Duration{time.Second, 3 * time.Second, 5 * time.Second}

	return cfg, nil
}

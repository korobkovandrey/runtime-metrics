package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/korobkovandrey/runtime-metrics/internal/server/db"
	"github.com/korobkovandrey/runtime-metrics/internal/server/db/pgxdriver"
)

type Config struct {
	Addr            string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	Restore         bool   `env:"RESTORE"`
	StoreInterval   int64  `env:"STORE_INTERVAL"`
	ShutdownTimeout int64
}

func GetConfig() (*Config, error) {
	const (
		storeInterval   = 0
		shutdownTimeout = 5
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
		return cfg, fmt.Errorf("GetConfig: %w", err)
	}

	cfg.ShutdownTimeout = shutdownTimeout

	return cfg, nil
}

func (cfg *Config) GetDBConfig() *db.Config {
	dbCfg := &db.Config{}
	if cfg.DatabaseDSN != "" {
		dbCfg.PGXDriver = &pgxdriver.Config{DSN: cfg.DatabaseDSN}
	}
	return dbCfg
}

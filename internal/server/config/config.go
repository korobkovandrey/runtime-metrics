package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr            string `env:"ADDRESS"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	StoreInterval   int64  `env:"STORE_INTERVAL"`
}

func GetConfig() (*Config, error) {
	const (
		storeInterval = 0
	)
	cfg := &Config{}
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server host")
	flag.Int64Var(&cfg.StoreInterval, "i", storeInterval, "store interval")
	flag.StringVar(&cfg.FileStoragePath, "f", "storage.json", "file storage path")
	flag.BoolVar(&cfg.Restore, "r", true, "file storage path")

	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return cfg, fmt.Errorf("GetConfig: %w", err)
	}

	return cfg, nil
}

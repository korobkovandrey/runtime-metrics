package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr string `env:"ADDRESS"`
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server host")

	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return cfg, fmt.Errorf("GetConfig: %w", err)
	}

	return cfg, nil
}

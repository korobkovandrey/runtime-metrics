package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"

	"strings"
)

type Config struct {
	Addr       string `env:"ADDRESS"`
	UpdatePath string
	ValuePath  string
}

func GetConfig() (*Config, error) {
	cfg := &Config{}
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server host")
	flag.StringVar(&cfg.UpdatePath, "updatePath", "update", "update path")
	flag.StringVar(&cfg.ValuePath, "valuePath", "value", "value path")

	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return cfg, fmt.Errorf("GetConfig: %w", err)
	}

	cfg.UpdatePath = "/" + strings.Trim(cfg.UpdatePath, "/")
	cfg.ValuePath = "/" + strings.Trim(cfg.ValuePath, "/")
	return cfg, nil
}

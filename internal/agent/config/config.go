package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/sender"
)

type Config struct {
	Addr           string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	Sender         *sender.Config
}

func GetConfig() (*Config, error) {
	const (
		pollIntervalSeconds   = 2
		reportIntervalSeconds = 10
	)
	cfg := &Config{}
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server host")
	flag.IntVar(&cfg.PollInterval, "p", pollIntervalSeconds, "pollInterval in seconds")
	flag.IntVar(&cfg.ReportInterval, "r", reportIntervalSeconds, "reportInterval in seconds")

	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return cfg, fmt.Errorf("GetConfig: %w", err)
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

	baseURL := "http://" + cfg.Addr
	cfg.Sender = &sender.Config{
		UpdateURL:  baseURL + "/update/",
		UpdatesURL: baseURL + "/updates/",
	}
	return cfg, nil
}

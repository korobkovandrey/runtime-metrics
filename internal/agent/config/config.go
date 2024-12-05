package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr               string `env:"ADDRESS"`
	UpdateURL          string
	PollInterval       int `env:"POLL_INTERVAL"`
	ReportInterval     int `env:"REPORT_INTERVAL"`
	ReportWorkersCount int
	TimeoutCoefficient float64
}

const (
	pollIntervalSeconds   = 2
	reportIntervalSeconds = 10
	reportWorkersCount    = 1
	timeoutCoefficient    = 0.25
)

func GetConfig() (*Config, error) {
	cfg := &Config{}
	flag.StringVar(&cfg.Addr, "a", "localhost:8080", "server host")
	flag.IntVar(&cfg.PollInterval, "p", pollIntervalSeconds, "pollInterval in seconds")
	flag.IntVar(&cfg.ReportInterval, "r", reportIntervalSeconds, "reportInterval in seconds")
	flag.IntVar(&cfg.ReportWorkersCount, "w", reportWorkersCount, "reportWorkersCount")
	flag.Float64Var(&cfg.TimeoutCoefficient, "t", timeoutCoefficient, "timeoutCoefficient")

	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return cfg, fmt.Errorf("GetConfig: %w", err)
	}

	cfg.UpdateURL = "http://" + cfg.Addr + "/update/"

	return cfg, nil
}

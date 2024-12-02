package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"

	"strings"
)

type Config struct {
	Addr               string `env:"ADDRESS"`
	UpdateGaugeURL     string
	UpdateCounterURL   string
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
	flag.StringVar(&cfg.Addr, `a`, `localhost:8080`, `server host`)
	updatePath := flag.String(`updatePath`, `update`, `update path`)
	gaugePath := flag.String(`gaugePath`, `gauge`, `gauge path`)
	counterPath := flag.String(`counterPath`, `counter`, `counter path`)
	flag.IntVar(&cfg.PollInterval, `p`, pollIntervalSeconds, `pollInterval in seconds`)
	flag.IntVar(&cfg.ReportInterval, `r`, reportIntervalSeconds, `ReportInterval in seconds`)
	flag.IntVar(&cfg.ReportWorkersCount, `w`, reportWorkersCount, `ReportWorkersCount`)
	flag.Float64Var(&cfg.TimeoutCoefficient, `t`, timeoutCoefficient, `TimeoutCoefficient`)

	flag.Parse()

	err := env.Parse(cfg)
	if err != nil {
		return cfg, fmt.Errorf(`GetConfig: %w`, err)
	}

	updateURL := `http://` + cfg.Addr + `/` + strings.Trim(*updatePath, `/`) + `/`

	cfg.UpdateGaugeURL = updateURL + strings.Trim(*gaugePath, `/`) + `/`
	cfg.UpdateCounterURL = updateURL + strings.Trim(*counterPath, `/`) + `/`

	return cfg, nil
}

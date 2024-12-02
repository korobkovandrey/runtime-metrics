package config

import (
	"flag"

	"strings"
	"time"
)

type Config struct {
	UpdateGaugeURL     string
	UpdateCounterURL   string
	PollInterval       time.Duration
	ReportInterval     time.Duration
	ReportWorkersCount int
	TimeoutCoefficient float64
}

const (
	pollIntervalSeconds   = 2
	reportIntervalSeconds = 10
	reportWorkersCount    = 1
	timeoutCoefficient    = 0.25
)

func GetConfig() Config {
	cfg := Config{}
	addr := flag.String(`a`, `localhost:8080`, `server host`)
	updatePath := flag.String(`updatePath`, `update`, `update path`)
	gaugePath := flag.String(`gaugePath`, `value`, `gauge path`)
	counterPath := flag.String(`counterPath`, `counter`, `value path`)
	flag.DurationVar(&cfg.PollInterval, `p`, pollIntervalSeconds*time.Second, `pollInterval`)
	flag.DurationVar(&cfg.ReportInterval, `r`, reportIntervalSeconds*time.Second, `ReportInterval`)
	flag.IntVar(&cfg.ReportWorkersCount, `w`, reportWorkersCount, `ReportWorkersCount`)
	flag.Float64Var(&cfg.TimeoutCoefficient, `t`, timeoutCoefficient, `TimeoutCoefficient`)

	flag.Parse()

	updateURL := `http://` + *addr + `/` + strings.Trim(*updatePath, `/`) + `/`

	cfg.UpdateGaugeURL = updateURL + strings.Trim(*gaugePath, `/`) + `/`
	cfg.UpdateCounterURL = updateURL + strings.Trim(*counterPath, `/`) + `/`

	return cfg
}

package main

import (
	"github.com/korobkovandrey/runtime-metrics/internal/agent"

	"log"
	"time"
)

const (
	serverUpdateBaseURL    = `http://localhost:8080/update/`
	gaugePath              = `gauge/`
	counterPath            = `counter/`
	pollIntervalSeconds    = 2
	reportIntervalSeconds  = 10
	reportWorkersCount     = 1
	HTTPTimeoutCoefficient = 0.25
)

func main() {
	if err := agent.New(&agent.Config{
		UpdateGaugeURL:         serverUpdateBaseURL + gaugePath,
		UpdateCounterURL:       serverUpdateBaseURL + counterPath,
		PollInterval:           pollIntervalSeconds * time.Second,
		ReportInterval:         reportIntervalSeconds * time.Second,
		ReportWorkersCount:     reportWorkersCount,
		HTTPTimeoutCoefficient: HTTPTimeoutCoefficient,
	}).Run(); err != nil {
		log.Fatal(err)
	}
}

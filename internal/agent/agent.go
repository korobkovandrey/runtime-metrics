package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/service"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

type Agent struct {
	gaugeSource *service.Source
	config      *config.Config
}

func sendRequest(client *http.Client, url string, contentType string, metrics *model.Metrics) error {
	var postBody io.Reader
	if metrics == nil {
		postBody = http.NoBody
	} else {
		m, err := json.Marshal(metrics)
		if err != nil {
			return fmt.Errorf("failed marshled metrics: %w", err)
		}
		postBody = bytes.NewBuffer(m)
	}
	response, err := client.Post(url, contentType, postBody)
	if err != nil {
		return fmt.Errorf("sendRequest: %w", err)
	}
	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				log.Printf("failed to close the response body: %v", err)
			}
		}()
	}
	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to read the response body: %w (status code %d)", err, response.StatusCode)
		}
		return fmt.Errorf("unexpected status code received: %d (body: %s)", response.StatusCode, string(body))
	}
	return nil
}

func New(cfg *config.Config) *Agent {
	return &Agent{
		gaugeSource: service.NewGaugeSource(),
		config:      cfg,
	}
}

func (a *Agent) Run() {
	go func() {
		tick := time.NewTicker(time.Duration(a.config.PollInterval) * time.Second)
		for ; ; <-tick.C {
			a.gaugeSource.Collect()
		}
	}()

	client := &http.Client{}
	var pollCount, pollCountDelta, sentPollCount int64
	var err error
	metricsPollCount := model.Metrics{
		Delta: &pollCountDelta,
		ID:    "PollCount",
		MType: "counter",
	}
	metricsGauge := model.Metrics{
		MType: "gauge",
	}
	for range time.Tick(time.Duration(a.config.ReportInterval) * time.Second) {
		pollCount = a.gaugeSource.GetPollCount()
		pollCountDelta = pollCount - sentPollCount
		err = sendRequest(client, a.config.UpdateURL, "application/json", &metricsPollCount)
		if err == nil {
			sentPollCount = pollCount
		} else {
			log.Printf("fail send %s: %v", "PollCount", err)
		}
		dataForSend := a.gaugeSource.GetDataForSend()
		for i, v := range dataForSend {
			metricsGauge.ID = i
			metricsGauge.Value = &v
			err = sendRequest(client, a.config.UpdateURL, "application/json", &metricsGauge)
			if err != nil {
				log.Printf("fail send %s: %v", i, err)
			}
		}
	}
}

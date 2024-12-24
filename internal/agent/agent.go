package agent

import (
	"bytes"
	"compress/gzip"
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

func sendRequest(client *http.Client, url string, contentType string, metric *model.Metric) error {
	var postBody io.Reader
	isGzipped := false
	if metric == nil {
		postBody = http.NoBody
	} else {
		m, err := json.Marshal(metric)
		if err != nil {
			return fmt.Errorf("failed marshled metric: %w", err)
		}
		buf := bytes.NewBuffer(nil)
		gz := gzip.NewWriter(buf)
		_, err = gz.Write(m)
		if err != nil {
			return fmt.Errorf("failed gzip write: %w", err)
		}
		err = gz.Close()
		if err != nil {
			return fmt.Errorf("failed gzip close: %w", err)
		}
		postBody = buf
		isGzipped = true
	}
	req, err := http.NewRequest(http.MethodPost, url, postBody)
	if err != nil {
		return fmt.Errorf("sendRequest: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept-Encoding", "gzip")
	if isGzipped {
		req.Header.Set("Content-Encoding", "gzip")
	}

	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed send request: %w", err)
	}
	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				log.Printf("failed to close the response body: %v", err)
			}
		}()
	}
	if response.StatusCode != http.StatusOK {
		var bodyReader io.ReadCloser

		if response.Header.Get("Content-Encoding") == "gzip" {
			bodyReader, err = gzip.NewReader(response.Body)
			if err != nil {
				return fmt.Errorf("failed gzip reader: %w", err)
			}
		} else {
			bodyReader = response.Body
		}
		body, err := io.ReadAll(bodyReader)
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
	metricPollCount := model.Metric{
		Delta: &pollCountDelta,
		ID:    "PollCount",
		MType: "counter",
	}
	metricGauge := model.Metric{
		MType: "gauge",
	}
	for range time.Tick(time.Duration(a.config.ReportInterval) * time.Second) {
		pollCount = a.gaugeSource.GetPollCount()
		pollCountDelta = pollCount - sentPollCount
		err = sendRequest(client, a.config.UpdateURL, "application/json", &metricPollCount)
		if err == nil {
			sentPollCount = pollCount
		} else {
			log.Printf("fail send %s: %v", "PollCount", err)
		}
		dataForSend := a.gaugeSource.GetDataForSend()
		for i, v := range dataForSend {
			metricGauge.ID = i
			metricGauge.Value = &v
			err = sendRequest(client, a.config.UpdateURL, "application/json", &metricGauge)
			if err != nil {
				log.Printf("fail send %s: %v", i, err)
			}
		}
	}
}

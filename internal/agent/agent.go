package agent

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/service"
)

type Agent struct {
	gaugeSource *service.Source
	config      *config.Config
}

func sendRequest(client *http.Client, url string) error {
	response, err := client.Post(url, "text/plain", http.NoBody)
	if err != nil {
		return fmt.Errorf("sendRequest: %w", err)
	}
	if response != nil {
		defer func() {
			err := response.Body.Close()
			log.Printf("failed to close the response body: %v", err)
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
	client := &http.Client{}
	go func(client *http.Client) {
		pollInterval := time.Duration(a.config.PollInterval) * time.Second
		var err error
		for {
			a.gaugeSource.Collect()
			err = sendRequest(client, a.config.UpdateURL+"counter/PollCount/1")
			if err != nil {
				log.Printf("fail send PollCount: %v", err)
			}
			time.Sleep(pollInterval)
		}
	}(client)

	for range time.Tick(time.Duration(a.config.ReportInterval) * time.Second) {
		dataForSend := a.gaugeSource.GetDataForSend()
		var err error
		for i, v := range dataForSend {
			err = sendRequest(client, a.config.UpdateURL+"gauge/"+i+"/"+v)
			if err != nil {
				log.Printf("fail send %s: %v", i, err)
			}
		}
	}
}

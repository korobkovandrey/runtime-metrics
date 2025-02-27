package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/service"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

type Agent struct {
	gaugeSource *service.Source
	config      *config.Config
	l           *logging.ZapLogger
	client      *http.Client
}

func (a *Agent) sendMetric(ctx context.Context, metric model.Metric) error {
	var postBody io.Reader
	const errMsg = "sendMetric: %w"
	m, err := json.Marshal(metric)
	if err == nil {
		buf := bytes.NewBuffer(nil)
		gz := gzip.NewWriter(buf)
		_, err = gz.Write(m)
		if err == nil {
			err = gz.Close()
			postBody = buf
		}
	}
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	req, err := http.NewRequest(http.MethodPost, a.config.UpdateURL, postBody)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")

	response, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf(errMsg, err)
	}
	if response != nil {
		defer func() {
			if err := response.Body.Close(); err != nil {
				a.l.ErrorCtx(ctx, "failed to close the response body", zap.Error(err))
			}
		}()
	}
	if response.StatusCode != http.StatusOK {
		var bodyReader io.ReadCloser

		if response.Header.Get("Content-Encoding") == "gzip" {
			bodyReader, err = gzip.NewReader(response.Body)
			if err != nil {
				return fmt.Errorf(errMsg, err)
			}
			defer func(bodyReader io.ReadCloser) {
				if err := bodyReader.Close(); err != nil {
					a.l.ErrorCtx(ctx, "failed to close the gzip reader", zap.Error(err))
				}
			}(bodyReader)
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

func New(cfg *config.Config, l *logging.ZapLogger) *Agent {
	return &Agent{
		gaugeSource: service.NewGaugeSource(),
		config:      cfg,
		l:           l,
		client:      &http.Client{},
	}
}

func (a *Agent) Run() {
	go func() {
		tick := time.NewTicker(time.Duration(a.config.PollInterval) * time.Second)
		for ; ; <-tick.C {
			a.gaugeSource.Collect()
		}
	}()

	var pollCount, pollCountDelta, sentPollCount int64
	var err error
	metricPollCount := model.Metric{
		Delta: &pollCountDelta,
		ID:    "PollCount",
		MType: model.TypeCounter,
	}
	metricGauge := model.Metric{
		MType: model.TypeGauge,
	}
	ctx := context.Background()

	for range time.Tick(time.Duration(a.config.ReportInterval) * time.Second) {
		pollCount = a.gaugeSource.GetPollCount()
		pollCountDelta = pollCount - sentPollCount
		err = a.sendMetric(a.l.WithContextFields(ctx, zap.Int64("PollCount", pollCountDelta)), metricPollCount)
		if err == nil {
			sentPollCount = pollCount
		} else {
			a.l.ErrorCtx(ctx, "fail send PollCount", zap.Error(err))
		}
		dataForSend := a.gaugeSource.GetDataForSend()
		for i, v := range dataForSend {
			metricGauge.ID = i
			metricGauge.Value = &v
			err = a.sendMetric(a.l.WithContextFields(ctx, zap.Float64(i, v)), metricGauge)
			if err != nil {
				a.l.ErrorCtx(ctx, "fail send "+i, zap.Error(err))
			}
		}
	}
}

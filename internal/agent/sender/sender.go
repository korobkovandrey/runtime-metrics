package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"syscall"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

type Config struct {
	UpdateURL   string
	UpdatesURL  string
	Timeout     time.Duration
	RetryDelays []time.Duration
}

type Sender struct {
	cfg    *Config
	l      *logging.ZapLogger
	client *http.Client
}

func New(cfg *Config, l *logging.ZapLogger) *Sender {
	return &Sender{cfg: cfg, l: l, client: &http.Client{
		Timeout: cfg.Timeout,
	}}
}

func (s *Sender) SendMetric(ctx context.Context, m model.Metric) error {
	_, err := s.postData(ctx, s.cfg.UpdateURL, m)
	if err != nil {
		return fmt.Errorf("failed to send metric: %w", err)
	}
	return nil
}

func (s *Sender) SendMetrics(ctx context.Context, ms []*model.Metric) ([]*model.Metric, error) {
	res, err := s.postData(ctx, s.cfg.UpdatesURL, ms)
	if err != nil {
		return nil, fmt.Errorf("failed to send metric: %w", err)
	}

	var resMs []*model.Metric
	if err := json.Unmarshal(res, &resMs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return resMs, nil
}

func makeGzipBuffer(data any) (*bytes.Buffer, error) {
	if data == nil {
		return nil, nil
	}
	m, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	buf := bytes.NewBuffer(nil)
	gz := gzip.NewWriter(buf)
	if _, err := gz.Write(m); err != nil {
		return nil, fmt.Errorf("failed to gzip data: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}
	return buf, nil
}

func (s *Sender) getBodyFromResponse(ctx context.Context, response *http.Response) ([]byte, error) {
	var bodyReader io.ReadCloser
	if response.Header.Get("Content-Encoding") == "gzip" {
		var err error
		bodyReader, err = gzip.NewReader(response.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer func() {
			if err := bodyReader.Close(); err != nil {
				s.l.WarnCtx(ctx, "failed to close the gzip reader", zap.Error(err))
			}
		}()
	} else {
		bodyReader = response.Body
	}
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return body, nil
}

func (s *Sender) postData(ctx context.Context, url string, data any) ([]byte, error) {
	postBody, err := makeGzipBuffer(data)
	if err != nil {
		return nil, fmt.Errorf("failed to make gzip buffer: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, postBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")

	var response *http.Response
	for i := 0; ; i++ {
		response, err = s.client.Do(req)
		if i == len(s.cfg.RetryDelays) {
			break
		}
		if err == nil {
			if response.StatusCode >= http.StatusOK || response.StatusCode < http.StatusInternalServerError {
				break
			}
			if err := response.Body.Close(); err != nil {
				s.l.WarnCtx(ctx, "failed to close the response body", zap.Error(err))
			}
		} else if !errors.Is(err, syscall.ECONNREFUSED) {
			break
		}
		s.l.WarnCtx(ctx, "failed to send request, will retry", zap.Int("attempt", i+1), zap.Error(err))
		time.Sleep(s.cfg.RetryDelays[i])
	}
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			s.l.WarnCtx(ctx, "failed to close the response body", zap.Error(err))
		}
	}()

	body, err := s.getBodyFromResponse(ctx, response)
	if err != nil {
		return nil, fmt.Errorf("failed to read the response body: %w (status code %d)", err, response.StatusCode)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code received: %d (body: %s)", response.StatusCode, string(body))
	}
	return body, nil
}

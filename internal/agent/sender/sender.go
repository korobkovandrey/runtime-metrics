// Package sender contains the sender logic.
package sender

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
)

// Config contains the configuration for the sender.
type Config struct {
	UpdateURL   string
	UpdatesURL  string
	RetryDelays []time.Duration
	Key         []byte
	Timeout     time.Duration
	RateLimit   int
}

// Sender sends metrics to the server.
type Sender struct {
	cfg    *Config
	l      *logging.ZapLogger
	client *http.Client
}

// New creates a new sender.
func New(cfg *Config, l *logging.ZapLogger) *Sender {
	return &Sender{cfg: cfg, l: l, client: &http.Client{
		Timeout: cfg.Timeout,
	}}
}

// SendMetric sends a metric to the server.
func (s *Sender) SendMetric(ctx context.Context, m *model.Metric) error {
	if err := s.postData(ctx, s.cfg.UpdateURL, m); err != nil {
		return fmt.Errorf("failed to send metric: %w", err)
	}
	return nil
}

// SendBatchMetrics sends a batch of metrics to the server.
func (s *Sender) SendBatchMetrics(ctx context.Context, ms []*model.Metric) error {
	if err := s.postData(ctx, s.cfg.UpdatesURL, ms); err != nil {
		return fmt.Errorf("failed to send metric: %w", err)
	}
	return nil
}

// JobResult contains the result of a job.
type JobResult struct {
	*model.Metric
	Err error
}

// SendPoolMetrics sends metrics to the server in parallel.
func (s *Sender) SendPoolMetrics(ctx context.Context, numWorkers int, ms []*model.Metric) <-chan *JobResult {
	jobs := make(chan *model.Metric, len(ms))
	results := make(chan *JobResult, len(ms))
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for j := range jobs {
				if ctx.Err() != nil {
					break
				}
				results <- &JobResult{
					Metric: j,
					Err:    s.SendMetric(ctx, j),
				}
			}
		}()
	}
	for _, m := range ms {
		jobs <- m
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}

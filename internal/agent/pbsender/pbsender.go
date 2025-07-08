package pbsender

import (
	"context"
	"fmt"
	"sync"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/sender"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Addr          string
	RealIPAddress string
	Key           []byte
	RateLimit     int
}

type Sender struct {
	c    proto.MetricsServiceClient
	conn *grpc.ClientConn
	cfg  *Config
}

func New(cfg *Config) (*Sender, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(getRealIPInterceptor(cfg.RealIPAddress)),
	}
	if len(cfg.Key) > 0 {
		dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(getHashInterceptor(cfg.Key)))
	}
	conn, err := grpc.NewClient(
		cfg.Addr,
		dialOpts...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}
	return &Sender{
		c:    proto.NewMetricsServiceClient(conn),
		conn: conn,
		cfg:  cfg,
	}, nil
}

func (s *Sender) Close() error {
	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("failed to close grpc client: %w", err)
	}
	return nil
}

func (s *Sender) SendBatchMetrics(ctx context.Context, ms []*model.Metric) error {
	req := &proto.Metrics{Metrics: model.ModelMetricsToMetrics(ms)}
	if _, err := s.c.Updates(ctx, req); err != nil {
		return fmt.Errorf("failed to send batch metrics: %w", err)
	}
	return nil
}

func (s *Sender) SendPoolMetrics(ctx context.Context, ms []*model.Metric) <-chan *sender.JobResult {
	jobs := make(chan *model.Metric, len(ms))
	results := make(chan *sender.JobResult, len(ms))
	var wg sync.WaitGroup
	wg.Add(s.cfg.RateLimit)
	for i := 0; i < s.cfg.RateLimit; i++ {
		go func() {
			defer wg.Done()
			for j := range jobs {
				if ctx.Err() != nil {
					break
				}
				_, err := s.c.Update(ctx, model.ModelMetricToMetric(j))
				results <- &sender.JobResult{
					Metric: j,
					Err:    err,
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

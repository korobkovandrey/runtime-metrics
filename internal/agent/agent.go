// Package agent contains the agent logic.
//
// The agent is responsible for collecting runtime metrics, and sending them to the server.
package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/sender"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/service"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
)

// Run starts the agent.
func Run(ctx context.Context, cfg *config.Config, l *logging.ZapLogger) {
	source := service.NewSource()
	tickPoll := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer tickPoll.Stop()
	go func() {
		for ; ; <-tickPoll.C {
			if err := source.Collect(ctx); err != nil {
				l.ErrorCtx(ctx, fmt.Errorf("failed to collect metrics: %w", err).Error())
			}
		}
	}()
	sendClient := sender.New(cfg.Sender, l)
	tickReport := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer tickReport.Stop()
	go func() {
		for range tickReport.C {
			data, delta := source.Get()
			if len(data) == 0 {
				continue
			}
			if cfg.Batching {
				if err := sendClient.SendBatchMetrics(ctx, data); err == nil {
					source.Commit(delta)
				} else {
					l.ErrorCtx(ctx, fmt.Errorf("failed to send metrics: %w", err).Error())
				}
			} else {
				for result := range sendClient.SendPoolMetrics(ctx, cfg.RateLimit, data) {
					if result.Err != nil {
						l.ErrorCtx(ctx, fmt.Errorf("failed to send metric: %w", result.Err).Error())
					} else if result.Metric != nil && result.Metric.MType == model.TypeCounter && result.Metric.ID == "PollCount" {
						source.Commit(delta)
					}
				}
			}
		}
	}()
	<-ctx.Done()
}

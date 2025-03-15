package agent

import (
	"context"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/sender"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/service"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

type Agent struct {
	config      *config.Config
	l           *logging.ZapLogger
	sender      *sender.Sender
	gaugeSource *service.Source
}

func New(cfg *config.Config, l *logging.ZapLogger, s *sender.Sender) *Agent {
	return &Agent{
		sender:      s,
		gaugeSource: service.NewGaugeSource(),
		config:      cfg,
		l:           l,
	}
}

func (a *Agent) Run(ctx context.Context) {
	a.l.InfoCtx(ctx, "Agent run with config", zap.Any("config", a.config))
	tickPoll := time.NewTicker(time.Duration(a.config.PollInterval) * time.Second)
	go func() {
		for ; ; <-tickPoll.C {
			a.gaugeSource.Collect()
		}
	}()

	tickReport := time.NewTicker(time.Duration(a.config.ReportInterval) * time.Second)
	go func() {
		var pollCount, pollCountDelta, sentPollCount int64
		for range tickReport.C {
			dataForSend := a.gaugeSource.GetDataForSend()
			pollCount = a.gaugeSource.GetPollCount()
			pollCountDelta = pollCount - sentPollCount
			ms := make([]*model.Metric, len(dataForSend)+1)
			i := 0
			for id, v := range dataForSend {
				ms[i] = model.NewMetricGauge(id, v)
				i++
			}
			ms[i] = model.NewMetricCounter("PollCount", pollCountDelta)
			if _, err := a.sender.SendMetrics(ctx, ms); err != nil {
				a.l.ErrorCtx(ctx, "fail send", zap.Error(err))
			} else {
				sentPollCount = pollCount
			}
		}
	}()
	<-ctx.Done()
	tickPoll.Stop()
	tickReport.Stop()
}

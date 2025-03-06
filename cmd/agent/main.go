package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/korobkovandrey/runtime-metrics/internal/agent"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/sender"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"

	"log"
)

func main() {
	l, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Sync()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.GetConfig()
	if err != nil {
		l.FatalCtx(ctx, "failed to get config", zap.Error(err))
	}

	agent.New(cfg, l, sender.New(cfg.Sender, l)).Run(ctx)
}

package main

import (
	"context"

	"github.com/korobkovandrey/runtime-metrics/internal/agent"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"
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
	ctx := context.Background()
	cfg, err := config.GetConfig()
	if err != nil {
		l.FatalCtx(ctx, "failed to get config", zap.Error(err))
	}
	agent.New(cfg, l).Run()
}

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"

	"log"
	//nolint:gosec // G108
	_ "net/http/pprof"
)

func main() {
	l, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Sync()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.NewConfig()
	if err != nil {
		l.FatalCtx(ctx, "failed to get config", zap.Error(err))
	}

	l.InfoCtx(ctx, "Agent run with cfg", zap.Any("cfg", cfg))

	if cfg.PprofAddr == "" {
		agent.Run(ctx, cfg, l)
	} else {
		go agent.Run(ctx, cfg, l)
		server := &http.Server{
			Addr:              cfg.PprofAddr,
			ReadHeaderTimeout: 3 * time.Second,
		}
		if err = server.ListenAndServe(); err != nil {
			l.FatalCtx(ctx, fmt.Errorf("pprof server error: %w", err).Error())
		}
	}
}

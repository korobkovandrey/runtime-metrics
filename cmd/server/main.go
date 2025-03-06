package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/controller"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"github.com/korobkovandrey/runtime-metrics/internal/server/service"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
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

	repo, closer, pinger, err := repository.Factory(ctx, cfg, l)
	if err != nil {
		l.FatalCtx(ctx, "failed to start repository", zap.Error(err))
	}
	if closer != nil {
		defer func(closer repository.Closer) {
			l.InfoCtx(ctx, "Closing repository...")
			if err := closer.Close(); err != nil {
				l.ErrorCtx(ctx, "failed to close repository", zap.Error(err))
			}
		}(closer)
	}

	c := controller.NewController(cfg, service.NewService(repo), l).
		WithPinger(pinger)
	if err = c.ListenAndServe(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.FatalCtx(ctx, "failed to start server", zap.Error(err))
	}
}

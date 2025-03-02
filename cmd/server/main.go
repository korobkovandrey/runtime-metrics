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

	ctx, cancel := context.WithCancel(context.Background())
	go func(cancel context.CancelFunc) {
		defer cancel()
		stop := make(chan os.Signal, 1)
		defer close(stop)
		signal.Notify(stop, os.Interrupt)
		<-stop
	}(cancel)

	cfg, err := config.GetConfig()
	if err != nil {
		l.FatalCtx(ctx, "failed to get config", zap.Error(err))
	}

	r, err := repository.Factory(ctx, cfg, l)
	if err != nil {
		l.FatalCtx(ctx, "failed to start store", zap.Error(err))
	}
	defer func(r service.Repository) {
		l.InfoCtx(ctx, "Closing repository...")
		if err := r.Close(); err != nil {
			l.ErrorCtx(ctx, "failed to close repository", zap.Error(err))
		}
	}(r)

	c := controller.NewController(cfg, service.NewService(r), l)
	if err = c.ListenAndServe(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.FatalCtx(ctx, "failed to start server", zap.Error(err))
	}
}

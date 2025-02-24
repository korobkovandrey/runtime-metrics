package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"

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

	ctx := context.Background()
	cfg, err := config.GetConfig()
	if err != nil {
		l.FatalCtx(ctx, "failed to get config", zap.Error(err))
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}
	defer func() {
		cancel()
		wg.Wait()
	}()
	r, err := repository.Factory(ctxWithCancel, wg, cfg, l)
	if err != nil {
		l.FatalCtx(ctx, "failed to start store", zap.Error(err))
	}
	c := controller.NewController(cfg, service.NewService(r), l)
	if err = c.ServeHTTP(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.FatalCtx(ctx, "failed to start server", zap.Error(err))
	}
}

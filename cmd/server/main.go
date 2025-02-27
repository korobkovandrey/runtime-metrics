package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	ctxStopSignal, cancelStopSignal := context.WithCancel(ctx)
	go func(cancel context.CancelFunc) {
		defer cancel()
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
	}(cancelStopSignal)

	cfg, err := config.GetConfig()
	if err != nil {
		l.FatalCtx(ctxStopSignal, "failed to get config", zap.Error(err))
	}

	wg := &sync.WaitGroup{}
	ctxWg, cancelWg := context.WithCancel(ctx)
	defer func(cancel context.CancelFunc, wg *sync.WaitGroup) {
		cancel()
		wg.Wait()
	}(cancelWg, wg)
	r, err := repository.Factory(ctxWg, wg, cfg, l)
	if err != nil {
		l.FatalCtx(ctxWg, "failed to start store", zap.Error(err))
	}
	c := controller.NewController(cfg, service.NewService(r), l)
	if err = c.ServeHTTP(ctxStopSignal); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.FatalCtx(ctxStopSignal, "failed to start server", zap.Error(err))
	}
}

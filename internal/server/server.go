package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
)

func ListenAndServe(ctx context.Context, l *logging.ZapLogger, cfg config.Config, handler http.Handler) error {
	server := http.Server{
		Addr:              cfg.Addr,
		ErrorLog:          l.Std(),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func(ctx context.Context) {
		ctxWithoutCancel := context.WithoutCancel(ctx)
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(ctxWithoutCancel, cfg.ShutdownTimeout)
		defer cancel()
		l.InfoCtx(ctx, "Shutting down the HTTP server...")
		if err := server.Shutdown(ctx); err != nil {
			l.ErrorCtx(ctx, fmt.Errorf("failed to shutdown server: %w", err).Error())
		}
	}(ctx)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

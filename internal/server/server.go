// Package server contains the server logic.
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
)

// ListenAndServe starts the HTTP server.
func ListenAndServe(ctx context.Context, l *logging.ZapLogger,
	addr string, shutdownTimeout time.Duration, handler http.Handler) error {
	server := http.Server{
		Addr:              addr,
		ErrorLog:          l.Std(),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		ctxWithoutCancel := context.WithoutCancel(ctx)
		<-ctx.Done()
		shCtx, cancel := context.WithTimeout(ctxWithoutCancel, shutdownTimeout)
		defer cancel()
		l.InfoCtx(shCtx, "Shutting down the HTTP server...")
		if err := server.Shutdown(shCtx); err != nil {
			l.ErrorCtx(shCtx, fmt.Errorf("failed to shutdown server: %w", err).Error())
		}
	}()
	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

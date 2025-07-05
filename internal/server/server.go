// Package server contains the server logic.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	pb "github.com/korobkovandrey/runtime-metrics/internal/proto"
	"github.com/korobkovandrey/runtime-metrics/internal/server/interceptors/ilogger"
	"github.com/korobkovandrey/runtime-metrics/internal/server/interceptors/isign"
	"github.com/korobkovandrey/runtime-metrics/internal/server/interceptors/isubnet"
	"github.com/korobkovandrey/runtime-metrics/internal/server/pbservice"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"google.golang.org/grpc"
)

// ListenAndServeHTTP starts the HTTP server.
func ListenAndServeHTTP(ctx context.Context, l *logging.ZapLogger,
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
			l.ErrorCtx(shCtx, fmt.Errorf("failed to shutdown HTTP server: %w", err).Error())
		}
	}()
	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}
	return nil
}

// ListenAndServeGRPC starts the GRPC server.
func ListenAndServeGRPC(ctx context.Context, l *logging.ZapLogger,
	addr string, ipNet *net.IPNet, key string, ms *pbservice.MetricsService) error {
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			ilogger.Interceptor(l),
			isubnet.Interceptor(ipNet),
			isign.Interceptor([]byte(key)),
		),
	)
	pb.RegisterMetricsServiceServer(s, ms)
	serv, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	go func() {
		<-ctx.Done()
		l.InfoCtx(ctx, "Shutting down the GRPC server...")
		s.GracefulStop()
	}()
	if err = s.Serve(serv); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("failed to start GRPC server: %w", err)
	}
	return nil
}

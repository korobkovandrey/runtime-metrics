// Package main initializes and starts the server application.
// It sets up logging, configures the server, and handles graceful shutdowns.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/korobkovandrey/runtime-metrics/internal/server"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/factory"
	"github.com/korobkovandrey/runtime-metrics/internal/server/pbservice"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

//go:generate go run ../../tools/genversion

func main() {
	l, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Sync()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	cfg, err := config.NewConfig()
	if err != nil {
		l.FatalCtx(ctx, fmt.Errorf("failed to get config: %w", err).Error())
	}
	rep, err := factory.RepositoryFactory(ctx, cfg, l)
	if err != nil {
		l.FatalCtx(ctx, fmt.Errorf("failed to create repository: %w", err).Error())
	}
	if rCloser, ok := rep.(interface {
		Close() error
	}); ok {
		defer func() {
			l.InfoCtx(ctx, "Closing repository...")
			if rErr := rCloser.Close(); rErr != nil {
				l.ErrorCtx(ctx, fmt.Errorf("failed to close repository: %w", err).Error())
			}
		}()
	}
	h := server.NewHandler()
	if err = h.Configure(cfg, rep, l); err != nil {
		l.FatalCtx(ctx, fmt.Errorf("failed to configure handler: %w", err).Error())
	}
	printCfg := config.Config{
		Addr:                cfg.Addr,
		GRPSAddr:            cfg.GRPSAddr,
		ShutdownTimeout:     cfg.ShutdownTimeout,
		StoreInterval:       cfg.StoreInterval,
		DatabasePingTimeout: cfg.DatabasePingTimeout,
		RetryDelays:         cfg.RetryDelays,
		DatabaseDSN:         cfg.DatabaseDSN,
		FileStoragePath:     cfg.FileStoragePath,
		Restore:             cfg.Restore,
		Key:                 cfg.Key,
		Pprof:               cfg.Pprof,
		TrustedSubnet:       cfg.TrustedSubnet,
		IPNet:               cfg.IPNet,
		CryptoKey:           cfg.CryptoKey,
	}
	l.InfoCtx(ctx, "Start with config:", zap.Any("config", printCfg))
	g := new(errgroup.Group)
	if cfg.Addr != "" {
		l.InfoCtx(ctx, "HTTP server started on http://"+cfg.Addr+"/")
		g.Go(func() error {
			if gErr := server.ListenAndServeHTTP(ctx, l, cfg.Addr, cfg.ShutdownTimeout, h); gErr != nil && !errors.Is(gErr, http.ErrServerClosed) {
				return fmt.Errorf("failed to start server: %w", gErr)
			}
			return nil
		})
	}
	if cfg.GRPSAddr != "" {
		l.InfoCtx(ctx, "GRPC server started on "+cfg.GRPSAddr)
		g.Go(func() error {
			gErr := server.ListenAndServeGRPC(ctx, l, cfg.GRPSAddr, cfg.IPNet, cfg.Key, pbservice.NewMetricsService(rep))
			if gErr != nil && !errors.Is(gErr, http.ErrServerClosed) {
				return fmt.Errorf("failed to start GRPC server: %w", gErr)
			}
			return nil
		})
	}
	if err = g.Wait(); err != nil {
		l.FatalCtx(ctx, fmt.Errorf("failed server: %w", err).Error())
	}
}

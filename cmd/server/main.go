// Package main initializes and starts the server application.
// It sets up logging, configures the server, and handles graceful shutdowns.
package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/korobkovandrey/runtime-metrics/internal/server"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
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
	h := server.NewHandler()
	defer func() {
		l.InfoCtx(ctx, "Closing handler...")
		if err = h.Close(); err != nil {
			l.ErrorCtx(ctx, fmt.Errorf("failed to close handler: %w", err).Error())
		}
	}()
	if err = h.Configure(ctx, cfg, l); err != nil {
		l.FatalCtx(ctx, fmt.Errorf("failed to configure handler: %w", err).Error())
	}
	printCfg := config.Config{
		Addr:                cfg.Addr,
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
	if cfg.PrivateKey != nil {
		printCfg.PrivateKey = &rsa.PrivateKey{}
	}
	l.InfoCtx(ctx, "Server started on http://"+cfg.Addr+"/", zap.Any("config", printCfg))
	if err = server.ListenAndServe(ctx, l, cfg.Addr, cfg.ShutdownTimeout, h); err != nil && !errors.Is(err, http.ErrServerClosed) {
		l.FatalCtx(ctx, "failed to start server", zap.Error(err))
	}
}

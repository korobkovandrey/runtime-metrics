// Package main contains entry point for agent service.
// The agent service is simple web service which collect some runtime metrics
// and send them to server.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/pbsender"
	"github.com/korobkovandrey/runtime-metrics/internal/agent/sender"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"

	"log"
	//nolint:gosec // G108
	_ "net/http/pprof"
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
		l.FatalCtx(ctx, "failed to get config", zap.Error(err))
	}
	printCfg := config.Config{
		Addr:           cfg.Addr,
		Key:            cfg.Key,
		PprofAddr:      cfg.PprofAddr,
		PollInterval:   cfg.PollInterval,
		ReportInterval: cfg.ReportInterval,
		RateLimit:      cfg.RateLimit,
		Batching:       cfg.Batching,
		CryptoKey:      cfg.CryptoKey,
		RealIPAddress:  cfg.RealIPAddress,
	}
	if cfg.Sender != nil {
		printCfg.Sender = &sender.Config{
			UpdateURL:     cfg.Sender.UpdateURL,
			UpdatesURL:    cfg.Sender.UpdatesURL,
			RetryDelays:   cfg.Sender.RetryDelays,
			Key:           cfg.Sender.Key,
			Timeout:       cfg.Sender.Timeout,
			RateLimit:     cfg.Sender.RateLimit,
			RealIPAddress: cfg.Sender.RealIPAddress,
		}
	}
	l.InfoCtx(ctx, "Agent run with cfg", zap.Any("cfg", printCfg))

	var senderClient agent.SenderClient
	if cfg.Sender == nil {
		pbCfg := &pbsender.Config{
			Addr:          cfg.Addr,
			RealIPAddress: cfg.RealIPAddress,
			Key:           []byte(cfg.Key),
			RateLimit:     cfg.RateLimit,
		}
		l.InfoCtx(ctx, "Agent run with GRPC", zap.Any("cfg", pbCfg))
		senderClient, err = pbsender.New(pbCfg)
		if err != nil {
			l.FatalCtx(ctx, fmt.Errorf("failed to create pbsender: %w", err).Error())
		}
	} else {
		senderClient = sender.New(cfg.Sender, l)
	}

	if cfg.PprofAddr == "" {
		agent.Run(ctx, cfg, l, senderClient)
	} else {
		go agent.Run(ctx, cfg, l, senderClient)
		server := &http.Server{
			Addr:              cfg.PprofAddr,
			ReadHeaderTimeout: 3 * time.Second,
		}
		const shutdownTimeout = 5
		go func() {
			ctxWithoutCancel := context.WithoutCancel(ctx)
			<-ctx.Done()
			shCtx, cancel := context.WithTimeout(ctxWithoutCancel, shutdownTimeout*time.Second)
			defer cancel()
			if errSh := server.Shutdown(shCtx); errSh != nil {
				l.ErrorCtx(ctx, fmt.Errorf("failed to shutdown pprof server: %w", errSh).Error())
			}
		}()
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.FatalCtx(ctx, fmt.Errorf("pprof server error: %w", err).Error())
		}
	}
}

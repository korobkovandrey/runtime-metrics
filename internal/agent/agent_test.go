package agent

import (
	"context"
	"flag"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/config"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRun(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"test"}
	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
	cfg, err := config.NewConfig()
	require.NoError(t, err)
	cfg.PollInterval = 1
	cfg.ReportInterval = 2
	cfg.Batching = true
	flag.CommandLine = flag.NewFlagSet("", flag.ExitOnError)
	cfgPool, err := config.NewConfig()
	require.NoError(t, err)
	cfgPool.PollInterval = 1
	cfgPool.ReportInterval = 2
	cfgPool.Batching = false
	tests := []struct {
		cfg      *config.Config
		name     string
		duration time.Duration
	}{
		{
			name:     "batch mode",
			cfg:      cfg,
			duration: 3 * time.Second,
		},
		{
			name:     "pool mode",
			cfg:      cfgPool,
			duration: 3 * time.Second,
		},
	}

	server := &http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	}
	defer func() {
		_ = server.Shutdown(t.Context())
	}()
	go func() {
		_ = server.ListenAndServe()
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.duration)
			defer cancel()

			l, err := logging.NewZapLogger(zap.InfoLevel)
			require.NoError(t, err)
			defer l.Sync()

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.NotPanics(t, func() {
					Run(ctx, tt.cfg, l)
				})
			}()
			<-ctx.Done()
			wg.Wait()
		})
	}
}

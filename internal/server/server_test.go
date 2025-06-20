package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestListenAndServe(t *testing.T) {
	tests := []struct {
		name            string
		shutdownTimeout time.Duration
		duration        time.Duration
	}{
		{
			name:            "basic start and stop",
			shutdownTimeout: time.Second,
			duration:        2 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.duration)
			defer cancel()

			l, err := logging.NewZapLogger(zap.InfoLevel)
			require.NoError(t, err)
			defer l.Sync()

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			srv := httptest.NewServer(handler)
			srv.Close()

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err = ListenAndServe(ctx, l, srv.Listener.Addr().String(), tt.shutdownTimeout, handler); err != nil {
					assert.ErrorIs(t, err, http.ErrServerClosed)
				}
			}()
			time.Sleep(100 * time.Millisecond)

			client := &http.Client{Timeout: time.Second}
			//nolint:noctx // ignore
			resp, err := client.Get(srv.URL)
			assert.NoError(t, err)
			if err == nil {
				_ = resp.Body.Close()
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}
			<-ctx.Done()
			wg.Wait()
		})
	}
}

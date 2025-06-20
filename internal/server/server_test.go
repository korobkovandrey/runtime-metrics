package server

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestListenAndServe(t *testing.T) {
	l, err := logging.NewZapLogger(zap.InfoLevel)
	require.NoError(t, err)
	defer l.Sync()

	list, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	require.NoError(t, list.Close())
	addr := list.Addr().String()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		if errSrv := ListenAndServe(ctx, l, addr, 100*time.Millisecond, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})); errSrv != nil {
			assert.ErrorIs(t, errSrv, http.ErrServerClosed)
		}
	}()

	client := &http.Client{}
	//nolint:noctx // ignore
	resp, err := client.Get("http://" + addr + "/updates")
	if err == nil {
		_ = resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
	wg.Wait()
}

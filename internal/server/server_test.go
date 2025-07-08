package server

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/pbservice"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/korobkovandrey/runtime-metrics/pkg/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestListenAndServeHTTP(t *testing.T) {
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
		if errSrv := ListenAndServeHTTP(ctx, l, addr, 100*time.Millisecond, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestListenAndServeGRPC(t *testing.T) {
	l, err := logging.NewZapLogger(zap.InfoLevel)
	require.NoError(t, err)
	defer l.Sync()

	list, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	require.NoError(t, list.Close())
	addr := list.Addr().String()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(t.Context())
	rep := repository.NewMemStorage()
	go func() {
		defer wg.Done()
		assert.NoError(t, ListenAndServeGRPC(ctx, l, addr, nil, "", pbservice.NewMetricsService(rep)))
	}()
	time.Sleep(100 * time.Millisecond)

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.NewClient(
		addr,
		dialOpts...,
	)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, conn.Close())
	}()
	client := proto.NewMetricsServiceClient(conn)

	wantMetric := model.NewMetricCounter("test", 1)
	resp, err := client.Update(t.Context(), model.ModelMetricToMetric(wantMetric))
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	gotMetric, err := rep.Find(t.Context(), wantMetric.ToRequest())
	assert.NoError(t, err)
	assert.Equal(t, wantMetric, gotMetric)

	*wantMetric.Delta = 2
	resp, err = client.Updates(t.Context(), &proto.Metrics{Metrics: []*proto.Metric{model.ModelMetricToMetric(wantMetric)}})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	*wantMetric.Delta = 3
	gotMetric, err = rep.Find(t.Context(), wantMetric.ToRequest())
	assert.NoError(t, err)
	assert.Equal(t, wantMetric, gotMetric)

	cancel()
	wg.Wait()
}

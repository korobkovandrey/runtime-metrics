package pbsender

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/agent/sender"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

type mockMetricsServiceServer struct {
	proto.UnimplementedMetricsServiceServer
	updateErr  error
	updatesErr error
}

func (m *mockMetricsServiceServer) Update(_ context.Context, _ *proto.Metric) (*proto.Response, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	return &proto.Response{}, nil
}

func (m *mockMetricsServiceServer) Updates(_ context.Context, _ *proto.Metrics) (*proto.Response, error) {
	if m.updatesErr != nil {
		return nil, m.updatesErr
	}
	return &proto.Response{}, nil
}

func setupTestServer(t *testing.T, mock *mockMetricsServiceServer) (*grpc.Server, *bufconn.Listener) {
	listener := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	proto.RegisterMetricsServiceServer(s, mock)
	go func() {
		if err := s.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			t.Errorf("Server failed: %v", err)
		}
	}()
	return s, listener
}

func dialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestNew(t *testing.T) {
	s, err := New(&Config{
		Addr:          "localhost:0",
		RealIPAddress: "192.168.1.1",
		Key:           []byte("testkey"),
		RateLimit:     2,
	})
	require.NoError(t, err)
	assert.NotNil(t, s.c)
	assert.NotNil(t, s.conn)
	assert.NoError(t, s.Close())
}

func TestSender_SendBatchMetrics(t *testing.T) {
	mock := &mockMetricsServiceServer{}
	server, listener := setupTestServer(t, mock)
	defer server.Stop()
	cfg := &Config{
		Addr:          "localhost",
		RealIPAddress: "192.168.1.1",
		RateLimit:     2,
	}
	tests := []struct {
		updatesErr error
		name       string
		errMsg     string
		metrics    []*model.Metric
		wantErr    bool
	}{
		{
			name: "successful batch send",
			metrics: []*model.Metric{
				model.NewMetricCounter("counter1", 65),
				model.NewMetricGauge("gauge1", 12.34),
			},
			wantErr: false,
		},
		{
			name:    "empty metrics",
			metrics: []*model.Metric{},
			wantErr: false,
		},
		{
			name:       "server error",
			metrics:    []*model.Metric{model.NewMetricCounter("counter1", 65)},
			updatesErr: status.Error(codes.Internal, "server error"),
			wantErr:    true,
			errMsg:     "failed to send batch metrics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.updatesErr = tt.updatesErr
			s, err := New(cfg)
			require.NoError(t, err)
			defer func() {
				assert.NoError(t, s.Close())
			}()
			conn, err := grpc.NewClient(
				cfg.Addr,
				[]grpc.DialOption{
					grpc.WithContextDialer(dialer(listener)),
					grpc.WithTransportCredentials(insecure.NewCredentials()),
					grpc.WithUnaryInterceptor(getRealIPInterceptor(cfg.RealIPAddress)),
				}...,
			)
			require.NoError(t, err)
			s.conn = conn
			s.c = proto.NewMetricsServiceClient(conn)
			err = s.SendBatchMetrics(t.Context(), tt.metrics)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestSender_SendPoolMetrics(t *testing.T) {
	mock := &mockMetricsServiceServer{}
	server, listener := setupTestServer(t, mock)
	defer server.Stop()
	cfg := &Config{
		Addr:          "localhost",
		Key:           []byte("testkey"),
		RealIPAddress: "192.168.1.1",
		RateLimit:     2,
	}
	tests := []struct {
		updateErr error
		ctx       func() context.Context
		name      string
		errMsg    string
		metrics   []*model.Metric
		wantErr   bool
	}{
		{
			name: "successful pool send",
			metrics: []*model.Metric{
				model.NewMetricCounter("counter1", 65),
				model.NewMetricGauge("gauge1", 12.34),
			},
			ctx: func() context.Context {
				return t.Context()
			},
			wantErr: false,
		},
		{
			name:    "empty metrics",
			metrics: []*model.Metric{},
			ctx: func() context.Context {
				return t.Context()
			},
			wantErr: false,
		},
		{
			name: "server error",
			metrics: []*model.Metric{
				model.NewMetricCounter("counter1", 65),
			},
			updateErr: status.Error(codes.Internal, "server error"),
			ctx: func() context.Context {
				return t.Context()
			},
			wantErr: true,
			errMsg:  "server error",
		},
		{
			name: "context canceled",
			metrics: []*model.Metric{
				model.NewMetricCounter("counter1", 65),
			},
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				return ctx
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.updateErr = tt.updateErr
			s, err := New(cfg)
			require.NoError(t, err)
			defer func() {
				assert.NoError(t, s.Close())
			}()
			conn, err := grpc.NewClient(
				cfg.Addr,
				[]grpc.DialOption{
					grpc.WithContextDialer(dialer(listener)),
					grpc.WithTransportCredentials(insecure.NewCredentials()),
					grpc.WithUnaryInterceptor(getRealIPInterceptor(cfg.RealIPAddress)),
				}...,
			)
			require.NoError(t, err)
			s.conn = conn
			s.c = proto.NewMetricsServiceClient(conn)
			ctx := tt.ctx()
			results := s.SendPoolMetrics(ctx, tt.metrics)
			var wg sync.WaitGroup
			collectedResults := make([]*sender.JobResult, 0, len(tt.metrics))
			wg.Add(1)
			go func() {
				defer wg.Done()
				for r := range results {
					collectedResults = append(collectedResults, r)
				}
			}()
			wg.Wait()
			if tt.wantErr {
				for _, r := range collectedResults {
					require.Error(t, r.Err)
					assert.Contains(t, r.Err.Error(), tt.errMsg)
				}
				return
			}
			if len(tt.metrics) == 0 || ctx.Err() != nil {
				assert.Empty(t, collectedResults)
				return
			}
			assert.Equal(t, len(tt.metrics), len(collectedResults))
			for _, r := range collectedResults {
				assert.NoError(t, r.Err)
				assert.Contains(t, tt.metrics, r.Metric)
			}
		})
	}
}

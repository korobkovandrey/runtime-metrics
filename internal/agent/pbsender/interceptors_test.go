package pbsender

import (
	"context"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/pkg/sign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func TestGetRealIPInterceptor(t *testing.T) {
	ctx := context.Background()
	method := "/runtime_metrics.MetricsService/Update"
	realIP := "192.168.1.100"
	req := "test-request"
	reply := "test-response"
	cc := &grpc.ClientConn{}
	interceptor := getRealIPInterceptor(realIP)
	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		assert.True(t, ok)
		ip := md.Get("XRealIP")
		assert.Len(t, ip, 1)
		assert.Equal(t, realIP, ip[0])
		return nil
	}
	err := interceptor(ctx, method, req, reply, cc, invoker)
	assert.NoError(t, err)
}

func TestGetHashInterceptor(t *testing.T) {
	t.Run("NonProtoMessage", func(t *testing.T) {
		key := []byte("test-key")
		cc := &grpc.ClientConn{}
		interceptor := getHashInterceptor(key)
		nonProtoReq := "not-a-proto-message"
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			t.Fatal("Invoker should not be called")
			return nil
		}
		reply := "test-response"
		err := interceptor(t.Context(), "/runtime_metrics.MetricsService/Update", nonProtoReq, reply, cc, invoker)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request is not a proto message")
	})
	t.Run("ValidSignature", func(t *testing.T) {
		key := []byte("test-key")
		cc := &grpc.ClientConn{}
		interceptor := getHashInterceptor(key)
		m := model.NewMetricCounter("test-metric", 1)
		req := model.ModelMetricToMetric(m)
		dataBytes, err := proto.Marshal(req)
		require.NoError(t, err)
		expectedHash := sign.MakeToString(dataBytes, key)
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			assert.True(t, ok)
			hash := md.Get("HashSHA256")
			assert.Len(t, hash, 1)
			assert.Equal(t, expectedHash, hash[0])
			return nil
		}
		reply := "test-response"
		err = interceptor(t.Context(), "/runtime_metrics.MetricsService/Update", req, reply, cc, invoker)
		assert.NoError(t, err)
	})

	t.Run("EmptyKey", func(t *testing.T) {
		cc := &grpc.ClientConn{}
		interceptor := getHashInterceptor(nil)
		m := model.NewMetricCounter("test-metric", 1)
		req := model.ModelMetricToMetric(m)
		dataBytes, err := proto.Marshal(req)
		require.NoError(t, err)
		expectedHash := sign.MakeToString(dataBytes, nil)
		invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			md, ok := metadata.FromOutgoingContext(ctx)
			assert.True(t, ok)
			hash := md.Get("HashSHA256")
			assert.Len(t, hash, 1)
			assert.Equal(t, expectedHash, hash[0])
			return nil
		}
		reply := "test-response"
		err = interceptor(t.Context(), "/runtime_metrics.MetricsService/Update", req, reply, cc, invoker)
		assert.NoError(t, err)
	})
}

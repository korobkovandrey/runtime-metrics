package isign

import (
	"context"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/pkg/sign"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestInterceptor(t *testing.T) {
	t.Run("EmptyKey", func(t *testing.T) {
		interceptor := Interceptor(nil)
		expectedResponse := "test-response"
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return expectedResponse, nil
		}
		m := model.NewMetricCounter("test-metric", 1)
		req := model.ModelMetricToMetric(m)
		resp, err := interceptor(t.Context(), req, &grpc.UnaryServerInfo{FullMethod: "/runtime_metrics.MetricsService/Update"}, handler)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
	})

	t.Run("MissingMetadata", func(t *testing.T) {
		key := []byte("test-key")
		interceptor := Interceptor(key)
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "test-response", nil
		}
		m := model.NewMetricCounter("test-metric", 1)
		req := model.ModelMetricToMetric(m)
		resp, err := interceptor(t.Context(), req, &grpc.UnaryServerInfo{FullMethod: "/runtime_metrics.MetricsService/Update"}, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "missing HashSHA256")
	})

	t.Run("MissingHashSHA256", func(t *testing.T) {
		key := []byte("test-key")
		interceptor := Interceptor(key)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.New(map[string]string{}))
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "test-response", nil
		}
		m := model.NewMetricCounter("test-metric", 1)
		req := model.ModelMetricToMetric(m)
		resp, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/runtime_metrics.MetricsService/Update"}, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "missing HashSHA256")
	})

	t.Run("InvalidHashSHA256", func(t *testing.T) {
		key := []byte("test-key")
		interceptor := Interceptor(key)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.New(map[string]string{
			"HashSHA256": "invalid-hash",
		}))
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "test-response", nil
		}
		m := model.NewMetricCounter("test-metric", 1)
		req := model.ModelMetricToMetric(m)
		resp, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/runtime_metrics.MetricsService/Update"}, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "failed to decode HashSHA256")
	})

	t.Run("NonProtoMessage", func(t *testing.T) {
		key := []byte("test-key")
		interceptor := Interceptor(key)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.New(map[string]string{
			"HashSHA256": sign.MakeToString([]byte("data"), key),
		}))
		nonProtoReq := "not-a-proto-message"
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "test-response", nil
		}
		resp, err := interceptor(ctx, nonProtoReq, &grpc.UnaryServerInfo{FullMethod: "/runtime_metrics.MetricsService/Update"}, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
		assert.Contains(t, err.Error(), "request is not a proto message")
	})

	t.Run("InvalidSignature", func(t *testing.T) {
		key := []byte("test-key")
		interceptor := Interceptor(key)
		m := model.NewMetricCounter("test-metric", 1)
		req := model.ModelMetricToMetric(m)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.New(map[string]string{
			"HashSHA256": sign.MakeToString([]byte("invalid-data"), key),
		}))
		expectedResponse := "test-response"
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return expectedResponse, nil
		}
		resp, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/runtime_metrics.MetricsService/Update"}, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
		assert.Contains(t, err.Error(), "invalid signature")
	})

	t.Run("ValidSignature", func(t *testing.T) {
		key := []byte("test-key")
		interceptor := Interceptor(key)
		m := model.NewMetricCounter("test-metric", 1)
		req := model.ModelMetricToMetric(m)
		dataBytes, err := proto.Marshal(req)
		require.NoError(t, err)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.New(map[string]string{
			"HashSHA256": sign.MakeToString(dataBytes, key),
		}))
		expectedResponse := "test-response"
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return expectedResponse, nil
		}
		resp, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: "/runtime_metrics.MetricsService/Update"}, handler)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
	})
}

package isubnet

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestInterceptor(t *testing.T) {
	_, testSubnet, _ := net.ParseCIDR("192.168.1.0/24")
	req := "test-request"
	fullMethod := "/test.Service/Method"

	t.Run("NilSubnet", func(t *testing.T) {
		interceptor := Interceptor(nil)
		expectedResponse := "test-response"
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return expectedResponse, nil
		}
		resp, err := interceptor(t.Context(), req, &grpc.UnaryServerInfo{FullMethod: fullMethod}, handler)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
	})

	t.Run("MissingMetadata", func(t *testing.T) {
		interceptor := Interceptor(testSubnet)
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "test-response", nil
		}
		resp, err := interceptor(t.Context(), req, &grpc.UnaryServerInfo{FullMethod: fullMethod}, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "missing XRealIP")
	})

	t.Run("MissingXRealIP", func(t *testing.T) {
		interceptor := Interceptor(testSubnet)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.New(map[string]string{}))
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "test-response", nil
		}
		resp, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: fullMethod}, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "missing XRealIP")
	})

	t.Run("InvalidIP", func(t *testing.T) {
		interceptor := Interceptor(testSubnet)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.New(map[string]string{
			"XRealIP": "invalid-ip",
		}))
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "test-response", nil
		}
		resp, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: fullMethod}, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
		assert.Contains(t, err.Error(), "invalid XRealIP")
	})

	t.Run("IPNotInSubnet", func(t *testing.T) {
		interceptor := Interceptor(testSubnet)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.New(map[string]string{
			"XRealIP": "10.0.0.1",
		}))
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "test-response", nil
		}
		resp, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: fullMethod}, handler)
		assert.Nil(t, resp)
		assert.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
		assert.Contains(t, err.Error(), "XRealIP is not in the subnet")
	})

	t.Run("ValidIPInSubnet", func(t *testing.T) {
		interceptor := Interceptor(testSubnet)
		ctx := metadata.NewIncomingContext(t.Context(), metadata.New(map[string]string{
			"XRealIP": "192.168.1.100",
		}))
		expectedResponse := "test-response"
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return expectedResponse, nil
		}
		resp, err := interceptor(ctx, req, &grpc.UnaryServerInfo{FullMethod: fullMethod}, handler)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, resp)
	})
}

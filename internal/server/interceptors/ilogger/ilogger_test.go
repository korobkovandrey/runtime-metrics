package ilogger

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// captureOutput captures stderr output from the given function
func captureOutput(t *testing.T, f func()) string {
	originalStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()
	f()
	_ = w.Close()
	<-done
	os.Stderr = originalStderr
	return buf.String()
}

func TestInterceptor(t *testing.T) {
	t.Run("SuccessfulHandler", func(t *testing.T) {
		req := "test-request"
		fullMethod := "/test.Service/Method"
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return "test-response", nil
		}
		output := captureOutput(t, func() {
			l, err := logging.NewZapLogger(zap.InfoLevel)
			require.NoError(t, err)
			defer l.Sync()
			resp, err := Interceptor(l)(t.Context(), req, &grpc.UnaryServerInfo{FullMethod: fullMethod}, handler)
			assert.NoError(t, err)
			assert.Equal(t, "test-response", resp)
		})
		assert.Contains(t, output, `"method":"`+fullMethod+`"`)
		assert.NotContains(t, output, "error")
	})

	t.Run("HandlerWithError", func(t *testing.T) {
		req := "test-request"
		fullMethod := "/test.Service/Method"
		expectedError := errors.New("handler error")
		//nolint:unparam // ignore
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			return nil, expectedError
		}
		output := captureOutput(t, func() {
			l, err := logging.NewZapLogger(zap.InfoLevel)
			require.NoError(t, err)
			defer l.Sync()
			resp, err := Interceptor(l)(t.Context(), req, &grpc.UnaryServerInfo{FullMethod: fullMethod}, handler)
			assert.Error(t, err)
			assert.Equal(t, expectedError, err)
			assert.Nil(t, resp)
		})
		assert.Contains(t, output, `"method":"`+fullMethod+`"`)
		assert.Contains(t, output, `"error":"handler error"`)
	})
}

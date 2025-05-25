// Package mlogger provides middleware for logging response data.
package mlogger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

type (
	// responseData contains response data.
	responseData struct {
		status int
		size   int
	}
	// loggingResponseWriter is a wrapper for http.ResponseWriter.
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write writes data to the response.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	if err != nil {
		return size, fmt.Errorf("failed to write response: %w", err)
	}
	return size, nil
}

// WriteHeader writes the status code to the response.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// key is a type for log message key.
type key string

// LogMessageKey is a key for log message.
const LogMessageKey key = "logMessage"

// RequestLogger returns a middleware for logging request data.
func RequestLogger(l *logging.ZapLogger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rd := &responseData{
				status: 0,
				size:   0,
			}
			h.ServeHTTP(&loggingResponseWriter{
				ResponseWriter: w,
				responseData:   rd,
			}, r)

			ctx := r.Context()
			msg := ""
			if m := ctx.Value(LogMessageKey); m != nil {
				msg, _ = m.(string)
			}
			fields := []zap.Field{
				zap.Int("status", rd.status),
				zap.String("method", r.Method), zap.String("uri", r.RequestURI),
				zap.Duration("duration", time.Since(start)), zap.Int("size", rd.size),
			}
			if rs := r.Header.Get("HashSHA256"); rs != "" {
				fields = append(fields, zap.String("sign", rs))
			}
			l.InfoCtx(ctx, msg, fields...)
		})
	}
}

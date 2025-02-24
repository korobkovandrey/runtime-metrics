package mlogger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	if err != nil {
		return size, fmt.Errorf("mlogger[loggingResponseWriter].Write: %w", err)
	}
	return size, nil
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func RequestLogger(logger *logging.ZapLogger) func(h http.Handler) http.Handler {
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

			logger.InfoCtx(
				r.Context(), fmt.Sprintf("%s %s %d", r.Method, r.RequestURI, rd.status),
				zap.Duration("duration", time.Since(start)), zap.Int("size", rd.size),
			)
		})
	}
}

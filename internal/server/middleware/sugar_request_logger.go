package middleware

import (
	"fmt"
	"net/http"
	"time"

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
		return size, fmt.Errorf("loggingResponseWriter: %w", err)
	}
	return size, nil
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func SugarRequestLogger(logger *zap.SugaredLogger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rd := &responseData{
				status: 0,
				size:   0,
			}
			lw := loggingResponseWriter{
				ResponseWriter: w,
				responseData:   rd,
			}
			h.ServeHTTP(&lw, r)

			logger.Infoln(
				r.Method, r.RequestURI, time.Since(start),
				"status", rd.status,
				"size", rd.size,
			)
		})
	}
}

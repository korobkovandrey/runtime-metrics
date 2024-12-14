package logger

import (
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

var Log zap.SugaredLogger

func Initialize() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("logger.Initialize: %w", err)
	}
	Log = *logger.Sugar()
	return nil
}

func WithLogging(h http.Handler) http.Handler {
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

		Log.Infoln(
			r.Method, r.RequestURI, time.Since(start),
			"status", rd.status,
			"size", rd.size,
		)
	})
}

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

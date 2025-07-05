package mlogger

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// captureOutput captures the output written to os.Stderr during the execution of f
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

func TestRequestLogger(t *testing.T) {
	tests := []struct {
		nextHandler   http.Handler
		wantLogFields map[string]interface{}
		name          string
		method        string
		uri           string
		headerHash    string
		ctxLogMessage string
		wantResponse  string
		wantStatus    int
	}{
		{
			name:          "successful request",
			method:        http.MethodGet,
			uri:           "/test",
			headerHash:    "",
			ctxLogMessage: "",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("Hello, World!"))
			}),
			wantStatus:   http.StatusOK,
			wantResponse: "Hello, World!",
			wantLogFields: map[string]interface{}{
				"level":   "INFO",
				"message": "",
				"status":  float64(200),
				"method":  "GET",
				"uri":     "/test",
				"size":    float64(13),
			},
		},
		{
			name:          "request with HashSHA256",
			method:        http.MethodPost,
			uri:           "/api",
			headerHash:    "abc123",
			ctxLogMessage: "",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			}),
			wantStatus:   http.StatusOK,
			wantResponse: "OK",
			wantLogFields: map[string]interface{}{
				"level":   "INFO",
				"message": "",
				"status":  float64(200),
				"method":  "POST",
				"uri":     "/api",
				"size":    float64(2),
				"sign":    "abc123",
			},
		},
		{
			name:          "request with log message",
			method:        http.MethodGet,
			uri:           "/test/log",
			headerHash:    "",
			ctxLogMessage: "test message",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("Logged"))
			}),
			wantStatus:   http.StatusOK,
			wantResponse: "Logged",
			wantLogFields: map[string]interface{}{
				"level":   "INFO",
				"message": "test message",
				"status":  float64(200),
				"method":  "GET",
				"uri":     "/test/log",
				"size":    float64(6),
			},
		},
		{
			name:          "no content response",
			method:        http.MethodGet,
			uri:           "/no-content",
			headerHash:    "",
			ctxLogMessage: "",
			nextHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			}),
			wantStatus:   http.StatusNoContent,
			wantResponse: "",
			wantLogFields: map[string]interface{}{
				"level":   "INFO",
				"message": "",
				"status":  float64(204),
				"method":  "GET",
				"uri":     "/no-content",
				"size":    float64(0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.uri, nil)
			if tt.headerHash != "" {
				req.Header.Set("HashSHA256", tt.headerHash)
			}
			w := httptest.NewRecorder()
			logOutput := captureOutput(t, func() {
				logger, err := logging.NewZapLogger(zap.InfoLevel)
				require.NoError(t, err)
				defer logger.Sync()
				if tt.ctxLogMessage != "" {
					ctx := context.WithValue(req.Context(), LogMessageKey, tt.ctxLogMessage)
					req = req.WithContext(ctx)
				}
				middleware := RequestLogger(logger)(tt.nextHandler)
				middleware.ServeHTTP(w, req)
			})
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantResponse, w.Body.String())
			assert.NotEmpty(t, logOutput)
			logLines := strings.Split(strings.TrimSpace(logOutput), "\n")
			var logEntry map[string]interface{}
			err := json.Unmarshal([]byte(logLines[len(logLines)-1]), &logEntry)
			require.NoError(t, err)
			for k, wantValue := range tt.wantLogFields {
				if k == "duration" {
					duration, ok := logEntry["duration"].(string)
					assert.True(t, ok)
					_, err := time.ParseDuration(duration)
					assert.NoError(t, err)
					continue
				}
				assert.Equal(t, wantValue, logEntry[k], "Log field %s mismatch", k)
			}
		})
	}
}

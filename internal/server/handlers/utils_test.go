package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mlogger"
	"github.com/stretchr/testify/assert"
)

type mockResponseWriter struct {
	*httptest.ResponseRecorder
	writeErr error
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.ResponseRecorder.Write(b)
}

func TestResponseMarshaled(t *testing.T) {
	tests := []struct {
		data            interface{}
		writer          http.ResponseWriter
		name            string
		wantBody        string
		wantContentType string
		wantLogMessage  string
		wantStatus      int
	}{
		{
			name:            "successful response",
			data:            map[string]string{"key": "value"},
			writer:          httptest.NewRecorder(),
			wantStatus:      http.StatusOK,
			wantBody:        `{"key":"value"}`,
			wantContentType: "application/json",
			wantLogMessage:  "",
		},
		{
			name:            "json marshal error",
			data:            make(chan int), // Невалидный тип для JSON
			writer:          httptest.NewRecorder(),
			wantStatus:      http.StatusInternalServerError,
			wantBody:        "Internal Server Error\n",
			wantContentType: "text/plain; charset=utf-8",
			wantLogMessage:  "failed response: json: unsupported type: chan int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			responseMarshaled(tt.data, tt.writer, req)
			w := tt.writer
			if mrw, ok := w.(*mockResponseWriter); ok {
				w = mrw.ResponseRecorder
			}
			recorder := w.(*httptest.ResponseRecorder)
			assert.Equal(t, tt.wantStatus, recorder.Code)
			assert.Equal(t, tt.wantBody, recorder.Body.String())
			if tt.wantContentType != "" {
				assert.Equal(t, tt.wantContentType, recorder.Header().Get("Content-Type"))
			} else {
				assert.Empty(t, recorder.Header().Get("Content-Type"))
			}
			logMsg, ok := req.Context().Value(mlogger.LogMessageKey).(string)
			if tt.wantLogMessage != "" {
				assert.True(t, ok)
				assert.Equal(t, tt.wantLogMessage, logMsg)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

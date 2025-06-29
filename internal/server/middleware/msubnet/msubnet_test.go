package msubnet

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mlogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	_, subnet, err := net.ParseCIDR("192.168.1.0/24")
	require.NoError(t, err)

	tests := []struct {
		subnet         *net.IPNet
		name           string
		headerIP       string
		wantLogMessage string
		wantStatus     int
		setHeader      bool
		wantNextCalled bool
	}{
		{
			name:           "valid IP in subnet",
			subnet:         subnet,
			headerIP:       "192.168.1.10",
			setHeader:      true,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantLogMessage: "",
		},
		{
			name:           "nil subnet",
			subnet:         nil,
			headerIP:       "192.168.1.10",
			setHeader:      true,
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
			wantLogMessage: "",
		},
		{
			name:           "missing X-Real-IP header",
			subnet:         subnet,
			headerIP:       "",
			setHeader:      false,
			wantStatus:     http.StatusForbidden,
			wantNextCalled: false,
			wantLogMessage: "missing X-Real-IP header",
		},
		{
			name:           "invalid IP in header",
			subnet:         subnet,
			headerIP:       "invalid",
			setHeader:      true,
			wantStatus:     http.StatusForbidden,
			wantNextCalled: false,
			wantLogMessage: "invalid X-Real-IP header",
		},
		{
			name:           "IP not in subnet",
			subnet:         subnet,
			headerIP:       "10.0.0.1",
			setHeader:      true,
			wantStatus:     http.StatusForbidden,
			wantNextCalled: false,
			wantLogMessage: "X-Real-IP is not in the subnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.setHeader {
				req.Header.Set("X-Real-IP", tt.headerIP)
			}
			w := httptest.NewRecorder()
			nextCalled := false
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})
			middleware := Middleware(tt.subnet)(nextHandler)
			middleware.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantNextCalled, nextCalled)

			if tt.wantLogMessage != "" {
				m, ok := req.Context().Value(mlogger.LogMessageKey).(string)
				assert.True(t, ok)
				assert.Equal(t, tt.wantLogMessage, m)
			} else {
				_, ok := req.Context().Value(mlogger.LogMessageKey).(string)
				assert.False(t, ok)
			}
		})
	}
}

package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/mocks"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestController_ping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	l, err := logging.NewZapLogger(zap.InfoLevel)
	require.NoError(t, err)
	cfg := &config.Config{}

	tests := []struct {
		name            string
		mockPingerSetup func(mockPinger *mocks.MockPinger)
		wantCode        int
	}{
		{
			name: "ping ok",
			mockPingerSetup: func(mockPinger *mocks.MockPinger) {
				mockPinger.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "ping without db",
			wantCode: http.StatusOK,
		},
		{
			name: "ping error",
			mockPingerSetup: func(mockPinger *mocks.MockPinger) {
				mockPinger.EXPECT().Ping(gomock.Any()).Return(errors.New("ping error")).Times(1)
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockService(ctrl)
			c := NewController(cfg, mockService, l)
			r := httptest.NewRequest(http.MethodGet, "/ping", http.NoBody)
			w := httptest.NewRecorder()

			if tt.mockPingerSetup != nil {
				mockPinger := mocks.NewMockPinger(ctrl)
				tt.mockPingerSetup(mockPinger)
				c.WithPinger(mockPinger)
			}

			c.ping(w, r)

			require.Equal(t, tt.wantCode, w.Code)
		})
	}
}

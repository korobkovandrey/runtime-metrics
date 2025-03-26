package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/server/handlers/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewPingHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		mockSetup func(*mocks.MockPinger)
		wantCode  int
	}{
		{
			name: "ping ok",
			mockSetup: func(s *mocks.MockPinger) {
				s.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "ping without db",
			wantCode: http.StatusOK,
		},
		{
			name: "ping error",
			mockSetup: func(s *mocks.MockPinger) {
				s.EXPECT().Ping(gomock.Any()).Return(errors.New("ping error")).Times(1)
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/ping", http.NoBody)
			w := httptest.NewRecorder()
			if tt.mockSetup != nil {
				s := mocks.NewMockPinger(ctrl)
				tt.mockSetup(s)
				NewPingHandler(s)(w, r)
			} else {
				NewPingHandler(nil)(w, r)
			}
			require.Equal(t, tt.wantCode, w.Code)
		})
	}
}

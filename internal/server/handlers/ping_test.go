package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewPing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name            string
		mockPingerSetup func(mockPinger *MockPinger)
		wantCode        int
	}{
		{
			name: "ping ok",
			mockPingerSetup: func(mockPinger *MockPinger) {
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
			mockPingerSetup: func(mockPinger *MockPinger) {
				mockPinger.EXPECT().Ping(gomock.Any()).Return(errors.New("ping error")).Times(1)
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/ping", http.NoBody)
			w := httptest.NewRecorder()
			var f func(w http.ResponseWriter, r *http.Request)
			if tt.mockPingerSetup != nil {
				mockPinger := NewMockPinger(ctrl)
				tt.mockPingerSetup(mockPinger)
				f = NewPing(mockPinger)
			} else {
				f = NewPing(nil)
			}
			f(w, r)
			require.Equal(t, tt.wantCode, w.Code)
		})
	}
}

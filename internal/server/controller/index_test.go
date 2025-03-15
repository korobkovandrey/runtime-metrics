package controller

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/mocks"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestController_indexFunc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	l, err := logging.NewZapLogger(zap.InfoLevel)
	require.NoError(t, err)
	cfg := &config.Config{}

	tests := []struct {
		name            string
		serviceResponse []*model.Metric
		serviceError    error
		wantCode        int
		containsStrings []string
	}{
		{
			name: "success",
			serviceResponse: []*model.Metric{
				model.NewMetricGauge("Alloc", 123.4),
				model.NewMetricCounter("PollCount", 10),
				model.NewMetricGauge("RandomValue", 12.55),
			},
			wantCode: http.StatusOK,
			containsStrings: []string{
				"Alloc", "123.4",
				"PollCount", "10",
				"RandomValue", "12.55",
			},
		},
		{
			name:         "error",
			serviceError: errors.New("error"),
			wantCode:     http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockService(ctrl)
			mockService.EXPECT().FindAll(gomock.Any()).Return(tt.serviceResponse, tt.serviceError)
			c := NewController(cfg, mockService, l)

			currentDir, err := os.Getwd()
			require.NoError(t, err)
			err = os.Chdir("../../..")
			require.NoError(t, err)
			handler, err := c.indexFunc()
			require.NoError(t, err)
			err = os.Chdir(currentDir)
			require.NoError(t, err)

			r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			w := httptest.NewRecorder()
			handler(w, r)

			require.Equal(t, tt.wantCode, w.Code)
			body := w.Body.String()
			for _, str := range tt.containsStrings {
				require.Contains(t, body, str)
			}
		})
	}
}

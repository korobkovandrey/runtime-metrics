package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/handlers/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewIndexHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		serviceError    error
		name            string
		serviceResponse []*model.Metric
		containsStrings []string
		wantCode        int
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
			s := mocks.NewMockAllFinder(ctrl)
			s.EXPECT().FindAll(gomock.Any()).Return(tt.serviceResponse, tt.serviceError)

			currentDir, err := os.Getwd()
			require.NoError(t, err)
			t.Chdir("../../..")
			handler, err := NewIndexHandler(s)
			require.NoError(t, err)
			t.Chdir(currentDir)

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

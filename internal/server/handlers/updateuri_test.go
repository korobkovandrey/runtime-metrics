package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/handlers/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewUpdateURIHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name            string
		pathValues      map[string]string
		mockSetup       func(*mocks.MockUpdater)
		wantCode        int
		containsStrings []string
	}{
		{
			name: "valid update",
			pathValues: map[string]string{
				"type":  model.TypeGauge,
				"name":  "test",
				"value": "1",
			},
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("test", 1)})).
					Return(model.NewMetricGauge("test", 1), nil)
			},
			wantCode: http.StatusOK,
		},
		{
			name: "invalid type",
			pathValues: map[string]string{
				"type":  "invalid",
				"name":  "test",
				"value": "1",
			},
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"type is not valid"},
		},
		{
			name: "invalid value",
			pathValues: map[string]string{
				"type":  model.TypeGauge,
				"name":  "test",
				"value": "invalid",
			},
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"value is not valid"},
		},
		{
			name: "not found",
			pathValues: map[string]string{
				"type":  model.TypeGauge,
				"name":  "not_found",
				"value": "1",
			},
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("not_found", 1)})).
					Return(nil, model.ErrMetricNotFound)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: "service error",
			pathValues: map[string]string{
				"type":  model.TypeGauge,
				"name":  "error",
				"value": "1",
			},
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("error", 1)})).
					Return(nil, errors.New("unexpected error"))
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewMockUpdater(ctrl)
			tt.mockSetup(s)
			target := "/update"
			for _, v := range tt.pathValues {
				target += "/" + v
			}
			r := httptest.NewRequest(http.MethodPost, target, http.NoBody)
			for i, v := range tt.pathValues {
				r.SetPathValue(i, v)
			}
			w := httptest.NewRecorder()
			NewUpdateURIHandler(s)(w, r)
			require.Equal(t, tt.wantCode, w.Code)
			body := w.Body.String()
			for _, str := range tt.containsStrings {
				require.Contains(t, body, str)
			}
		})
	}
}

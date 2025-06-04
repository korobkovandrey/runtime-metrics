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

func TestNewValueURIHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		pathValues      map[string]string
		mockSetup       func(*mocks.MockFinder)
		name            string
		wantBody        string
		containsStrings []string
		wantCode        int
	}{
		{
			name: "valid",
			pathValues: map[string]string{
				"type": model.TypeGauge,
				"name": "test",
			},
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().
					Find(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("test", 0)})).
					Return(model.NewMetricGauge("test", 1), nil)
			},
			wantBody: "1",
			wantCode: http.StatusOK,
		},
		{
			name: "invalid type",
			pathValues: map[string]string{
				"type": "invalid",
				"name": "test",
			},
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().Find(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"type is not valid"},
		},
		{
			name: "not found",
			pathValues: map[string]string{
				"type": model.TypeGauge,
				"name": "not_found",
			},
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().
					Find(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("not_found", 0)})).
					Return(nil, model.ErrMetricNotFound)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: "service error",
			pathValues: map[string]string{
				"type": model.TypeGauge,
				"name": "error",
			},
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().
					Find(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("error", 0)})).
					Return(nil, errors.New("unexpected error"))
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewMockFinder(ctrl)
			tt.mockSetup(s)
			target := "/update"
			for _, v := range tt.pathValues {
				target += "/" + v
			}
			r := httptest.NewRequest(http.MethodGet, target, http.NoBody)
			for i, v := range tt.pathValues {
				r.SetPathValue(i, v)
			}
			w := httptest.NewRecorder()
			NewValueURIHandler(s)(w, r)
			require.Equal(t, tt.wantCode, w.Code)
			body := w.Body.String()
			if tt.wantBody != "" {
				require.Equal(t, tt.wantBody, body)
			}
			for _, str := range tt.containsStrings {
				require.Contains(t, body, str)
			}
		})
	}
}

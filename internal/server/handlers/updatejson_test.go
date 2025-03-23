package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewUpdateJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name             string
		json             string
		mockUpdaterSetup func(mockUpdater *MockUpdater)
		wantCode         int
		wantJSON         string
		containsStrings  []string
	}{
		{
			name: "valid update",
			json: `{"type":"gauge","id":"test","value":12.34}`,
			mockUpdaterSetup: func(mockUpdater *MockUpdater) {
				mockUpdater.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("test", 12.34)})).
					Return(model.NewMetricGauge("test", 12.34), nil)
			},
			wantCode: http.StatusOK,
			wantJSON: `{"type":"gauge","id":"test","value":12.34}`,
		},
		{
			name: "invalid json",
			json: `invalid`,
			mockUpdaterSetup: func(mockUpdater *MockUpdater) {
				mockUpdater.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			wantCode: http.StatusBadRequest,
			wantJSON: "",
			containsStrings: []string{
				"Bad Request",
			},
		},
		{
			name: "service error",
			json: `{"type":"gauge","id":"error","value":12.34}`,
			mockUpdaterSetup: func(mockUpdater *MockUpdater) {
				mockUpdater.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("error", 12.34)})).
					Return(nil, errors.New("unexpected error"))
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUpdater := NewMockUpdater(ctrl)
			tt.mockUpdaterSetup(mockUpdater)
			f := NewUpdateJSON(mockUpdater)

			r := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(tt.json))
			w := httptest.NewRecorder()
			f(w, r)

			require.Equal(t, tt.wantCode, w.Code)
			body := w.Body.String()
			if tt.wantJSON != "" {
				require.JSONEq(t, tt.wantJSON, body)
			}
			for _, str := range tt.containsStrings {
				require.Contains(t, body, str)
			}
		})
	}
}

package controller

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/mocks"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestController_updateURI(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	l, err := logging.NewZapLogger(zap.InfoLevel)
	require.NoError(t, err)
	cfg := &config.Config{}

	tests := []struct {
		name             string
		pathValues       map[string]string
		mockServiceSetup func(mockService *mocks.MockService)
		wantCode         int
		containsStrings  []string
	}{
		{
			name: "valid update",
			pathValues: map[string]string{
				"type":  model.TypeGauge,
				"name":  "test",
				"value": "1",
			},
			mockServiceSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
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
			mockServiceSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
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
			mockServiceSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
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
			mockServiceSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
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
			mockServiceSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("error", 1)})).
					Return(nil, errors.New("unexpected error"))
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockService(ctrl)
			tt.mockServiceSetup(mockService)
			c := NewController(cfg, mockService, l)

			target := "/update"
			for _, v := range tt.pathValues {
				target += "/" + v
			}
			r := httptest.NewRequest(http.MethodPost, target, http.NoBody)
			for i, v := range tt.pathValues {
				r.SetPathValue(i, v)
			}
			w := httptest.NewRecorder()
			c.updateURI(w, r)

			require.Equal(t, tt.wantCode, w.Code)
			body := w.Body.String()
			for _, str := range tt.containsStrings {
				require.Contains(t, body, str)
			}
		})
	}
}

func TestController_updateJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	l, err := logging.NewZapLogger(zap.InfoLevel)
	require.NoError(t, err)
	cfg := &config.Config{}

	tests := []struct {
		name             string
		json             string
		mockServiceSetup func(mockService *mocks.MockService)
		wantCode         int
		wantJSON         string
		containsStrings  []string
	}{
		{
			name: "valid update",
			json: `{"type":"gauge","id":"test","value":12.34}`,
			mockServiceSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("test", 12.34)})).
					Return(model.NewMetricGauge("test", 12.34), nil)
			},
			wantCode: http.StatusOK,
			wantJSON: `{"type":"gauge","id":"test","value":12.34}`,
		},
		{
			name: "invalid json",
			json: `invalid`,
			mockServiceSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
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
			mockServiceSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("error", 12.34)})).
					Return(nil, errors.New("unexpected error"))
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockService(ctrl)
			tt.mockServiceSetup(mockService)
			c := NewController(cfg, mockService, l)

			r := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBufferString(tt.json))
			w := httptest.NewRecorder()
			c.updateJSON(w, r)

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

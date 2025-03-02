package controller

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/mocks"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestController_routes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	l, err := logging.NewZapLogger(zap.InfoLevel)
	require.NoError(t, err)
	cfg := &config.Config{}

	tests := []struct {
		name            string
		method          string
		url             string
		postBody        string
		mockSetup       func(mockService *mocks.MockService)
		wantCode        int
		wantContentType string
		wantJSON        string
		wantBody        string
		containsStrings []string
	}{
		// index group
		{
			name:   "index ok",
			method: http.MethodGet,
			url:    "/",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					FindAll().
					Return([]*model.Metric{
						model.NewMetricGauge("Alloc", 123.4),
						model.NewMetricCounter("PollCount", 10),
						model.NewMetricGauge("RandomValue", 12.55),
					}, nil)
			},
			wantCode:        http.StatusOK,
			wantContentType: "text/html; charset=utf-8",
			containsStrings: []string{
				"Alloc", "123.4",
				"PollCount", "10",
				"RandomValue", "12.55",
			},
		},
		{
			name:   "index service error",
			method: http.MethodGet,
			url:    "/",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					FindAll().
					Return(nil, errors.New("service error"))
			},
			wantCode: http.StatusInternalServerError,
		},
		// updateURI group
		{
			name:   "updateURI gauge valid",
			method: http.MethodPost,
			url:    "/update/gauge/test/1",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Update(gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("test", 1)})).
					Return(model.NewMetricGauge("test", 1), nil)
			},
			wantCode: http.StatusOK,
		},
		{
			name:   "updateURI counter valid",
			method: http.MethodPost,
			url:    "/update/counter/test/1",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Update(gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 1)})).
					Return(model.NewMetricCounter("test", 1), nil)
			},
			wantCode: http.StatusOK,
		},
		{
			name:   "updateURI service error",
			method: http.MethodPost,
			url:    "/update/counter/test/1",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Update(gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 1)})).
					Return(nil, errors.New("service error"))
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name:   "updateURI service error not found",
			method: http.MethodPost,
			url:    "/update/counter/test/1",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Update(gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 1)})).
					Return(nil, model.ErrMetricNotFound)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:   "updateURI fail value",
			method: http.MethodPost,
			url:    "/update/counter/test/fail",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Update(gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"value is not valid"},
		},
		{
			name:   "updateURI without value",
			method: http.MethodPost,
			url:    "/update/counter/test",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Update(gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"Value is required."},
		},
		{
			name:   "updateURI without name",
			method: http.MethodPost,
			url:    "/update/gauge",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Update(gomock.Any()).MaxTimes(0)
			},
			wantCode: http.StatusNotFound,
		},
		// updateJSON group
		{
			name:     "updateJSON valid gauge",
			method:   http.MethodPost,
			url:      "/update",
			postBody: `{"type":"gauge","id":"test","value":12.34}`,
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Update(gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("test", 12.34)})).
					Return(model.NewMetricGauge("test", 12.34), nil)
			},
			wantCode:        http.StatusOK,
			wantContentType: "application/json",
			wantJSON:        `{"type":"gauge","id":"test","value":12.34}`,
		},
		{
			name:     "updateJSON valid counter",
			method:   http.MethodPost,
			url:      "/update",
			postBody: `{"type":"counter","id":"test","delta":1}`,
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Update(gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 1)})).
					Return(model.NewMetricCounter("test", 1), nil)
			},
			wantCode:        http.StatusOK,
			wantContentType: "application/json",
			wantJSON:        `{"type":"counter","id":"test","delta":1}`,
		},
		{
			name:     "updateJSON invalid json",
			method:   http.MethodPost,
			url:      "/update",
			postBody: `invalid`,
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Update(gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"metric not found"},
		},
		{
			name:     "updateJSON service error",
			method:   http.MethodPost,
			url:      "/update",
			postBody: `{"type":"gauge","id":"error","value":12.34}`,
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Update(gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("error", 12.34)})).
					Return(nil, errors.New("service error"))
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "updateJSON fail type",
			method:   http.MethodPost,
			url:      "/update",
			postBody: `{"type":"fail","id":"test"}`,
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Update(gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"type is not valid"},
		},
		{
			name:     "updateJSON missing value",
			method:   http.MethodPost,
			url:      "/update",
			postBody: `{"type":"gauge","id":"test"}`,
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Update(gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"value is not valid"},
		},
		// valueURI group
		{
			name:   "valueURI gauge valid",
			method: http.MethodGet,
			url:    "/value/gauge/test",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Find(gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("test", 0)})).
					Return(model.NewMetricGauge("test", 10), nil)
			},
			wantCode:        http.StatusOK,
			wantContentType: "text/plain; charset=utf-8",
			wantBody:        "10",
		},
		{
			name:   "valueURI counter valid",
			method: http.MethodGet,
			url:    "/value/counter/test",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Find(gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 0)})).
					Return(model.NewMetricCounter("test", 111), nil)
			},
			wantCode:        http.StatusOK,
			wantContentType: "text/plain; charset=utf-8",
			wantBody:        "111",
		},
		{
			name:   "valueURI service error",
			method: http.MethodGet,
			url:    "/value/counter/test",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Find(gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 0)})).
					Return(nil, errors.New("service error"))
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name:   "valueURI not found metric",
			method: http.MethodGet,
			url:    "/value/counter/test",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Find(gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 0)})).
					Return(nil, model.ErrMetricNotFound)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name:   "valueURI fail type",
			method: http.MethodGet,
			url:    "/value/fail/name",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Find(gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"type is not valid"},
		},
		{
			name:   "valueURI without name",
			method: http.MethodGet,
			url:    "/value/counter",
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Find(gomock.Any()).MaxTimes(0)
			},
			wantCode: http.StatusNotFound,
		},
		// valueJSON group
		{
			name:     "valueJSON valid gauge",
			method:   http.MethodPost,
			url:      "/value",
			postBody: `{"type":"gauge","id":"test"}`,
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Find(gomock.Eq(&model.MetricRequest{Metric: &model.Metric{
						MType: model.TypeGauge,
						ID:    "test"}})).
					Return(model.NewMetricGauge("test", 12.34), nil)
			},
			wantCode:        http.StatusOK,
			wantContentType: "application/json",
			wantJSON:        `{"type":"gauge","id":"test","value":12.34}`,
		},
		{
			name:     "valueJSON valid counter",
			method:   http.MethodPost,
			url:      "/value",
			postBody: `{"type":"counter","id":"test"}`,
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Find(gomock.Eq(&model.MetricRequest{Metric: &model.Metric{
						MType: model.TypeCounter,
						ID:    "test"}})).
					Return(model.NewMetricCounter("test", 111), nil)
			},
			wantCode:        http.StatusOK,
			wantContentType: "application/json",
			wantJSON:        `{"type":"counter","id":"test","delta":111}`,
		},
		{
			name:     "valueJSON invalid json",
			method:   http.MethodPost,
			url:      "/value",
			postBody: `invalid`,
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Find(gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"metric not found"},
		},
		{
			name:     "valueJSON service error",
			method:   http.MethodPost,
			url:      "/value",
			postBody: `{"type":"gauge","id":"error"}`,
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().
					Find(gomock.Eq(&model.MetricRequest{Metric: &model.Metric{
						MType: model.TypeGauge,
						ID:    "error"}})).
					Return(nil, errors.New("service error"))
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name:     "valueJSON fail type",
			method:   http.MethodPost,
			url:      "/value",
			postBody: `{"type":"fail","id":"test"}`,
			mockSetup: func(mockService *mocks.MockService) {
				mockService.EXPECT().Find(gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"type is not valid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewMockService(ctrl)
			tt.mockSetup(mockService)
			c := NewController(cfg, mockService, l)
			currentDir, err := os.Getwd()
			require.NoError(t, err)
			err = os.Chdir("../../..")
			require.NoError(t, err)
			err = c.routes()
			require.NoError(t, err)
			err = os.Chdir(currentDir)
			require.NoError(t, err)
			ts := httptest.NewServer(c.r)
			defer ts.Close()

			var postBody io.Reader
			if tt.postBody == "" {
				postBody = http.NoBody
			} else {
				postBody = bytes.NewBufferString(tt.postBody)
			}

			gotBody, gotCode, gotContentType := testRequest(t, ts, tt.method, tt.url, postBody)
			gotBodyString := string(gotBody)

			require.Equal(t, tt.wantCode, gotCode)
			if tt.wantContentType != "" {
				assert.Equal(t, tt.wantContentType, gotContentType)
			}
			if tt.wantJSON != "" {
				assert.JSONEq(t, tt.wantJSON, gotBodyString)
			}
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, gotBodyString)
			}
			for _, str := range tt.containsStrings {
				require.Contains(t, gotBodyString, str)
			}
		})
	}
}

func testRequest(
	t *testing.T, ts *httptest.Server,
	method, path string, postBody io.Reader) (body []byte, statusCode int, contentType string) {
	t.Helper()
	req, err := http.NewRequest(method, ts.URL+path, postBody)
	require.NoError(t, err)
	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	statusCode = resp.StatusCode
	contentType = resp.Header.Get("Content-Type")
	return
}
